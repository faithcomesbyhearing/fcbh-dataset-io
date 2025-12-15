package revise_audio

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/stdio_exec"
)

// VoiceConverter interface for voice conversion operations
type VoiceConverter interface {
	// ConvertVoice converts source audio to target speaker's voice
	// referenceAudioPath is optional - if provided, uses it as reference for voice characteristics
	// Returns path to converted audio file
	ConvertVoice(sourceAudioPath string, targetSpeakerEmbeddingPath string, referenceAudioPath ...string) (string, *log.Status)
	
	// ExtractSpeakerEmbedding extracts speaker embedding from audio
	// Returns path to saved embedding file
	ExtractSpeakerEmbedding(audioPath string, outputPath string) (string, *log.Status)
}

// RVCAdapter implements VoiceConverter using FCBH RVC system
type RVCAdapter struct {
	ctx      context.Context
	config   VoiceConversionConfig
	vcPython *stdio_exec.StdioExec
}

// NewRVCAdapter creates a new RVC adapter
func NewRVCAdapter(ctx context.Context, config VoiceConversionConfig) *RVCAdapter {
	return &RVCAdapter{
		ctx:    ctx,
		config: config,
	}
}

// ConvertVoice converts source audio to target speaker's voice using RVC
// referenceAudioPath is the original audio file to use as reference for voice characteristics
func (r *RVCAdapter) ConvertVoice(sourceAudioPath string, targetSpeakerEmbeddingPath string, referenceAudioPath ...string) (string, *log.Status) {
	if status := r.EnsurePythonScriptReady(); status != nil {
		return "", status
	}
	
	// Create temporary output file
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, fmt.Sprintf("vc_output_%d.wav", os.Getpid()))
	
	// Prepare request
	request := map[string]interface{}{
		"mode":              "convert",
		"source":            sourceAudioPath,
		"target_embedding":  targetSpeakerEmbeddingPath,
		"output":            outputPath,
	}
	
	// Add reference audio path if provided
	if len(referenceAudioPath) > 0 && referenceAudioPath[0] != "" {
		request["reference_audio_path"] = referenceAudioPath[0]
	}
	
	if r.config.SampleRate > 0 {
		request["sample_rate"] = r.config.SampleRate
	}
	if r.config.F0Method != "" {
		request["f0_method"] = r.config.F0Method
	}
	if r.config.Device != "" {
		request["device"] = r.config.Device
	}
	
	if r.config.RVCModelPath != "" {
		request["rvc_model_path"] = r.config.RVCModelPath
	}
	
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", log.ErrorNoErr(r.ctx, 500, fmt.Sprintf("Failed to marshal request: %v", err))
	}
	
	// Send request to Python script
	responseJSON, status := r.vcPython.Process(string(requestJSON))
	if status != nil {
		return "", status
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		return "", log.ErrorNoErr(r.ctx, 500, fmt.Sprintf("Failed to parse response: %v", err))
	}
	
	if success, ok := response["success"].(bool); !ok || !success {
		errorMsg := "Unknown error"
		if errStr, ok := response["error"].(string); ok {
			errorMsg = errStr
		}
		return "", log.ErrorNoErr(r.ctx, 500, fmt.Sprintf("Voice conversion failed: %s", errorMsg))
	}
	
	outputPathStr, ok := response["output_path"].(string)
	if !ok || outputPathStr == "" {
		return "", log.ErrorNoErr(r.ctx, 500, "No output path in response")
	}
	
	return outputPathStr, nil
}

// ExtractSpeakerEmbedding extracts speaker embedding from audio
func (r *RVCAdapter) ExtractSpeakerEmbedding(audioPath string, outputPath string) (string, *log.Status) {
	if status := r.EnsurePythonScriptReady(); status != nil {
		return "", status
	}
	
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", log.ErrorNoErr(r.ctx, 500, fmt.Sprintf("Failed to create output directory: %v", err))
	}
	
	// Prepare request
	request := map[string]interface{}{
		"mode":   "extract",
		"source": audioPath,
		"output": outputPath,
	}
	
	if r.config.SampleRate > 0 {
		request["sample_rate"] = r.config.SampleRate
	}
	if r.config.F0Method != "" {
		request["f0_method"] = r.config.F0Method
	}
	if r.config.Device != "" {
		request["device"] = r.config.Device
	}
	
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", log.ErrorNoErr(r.ctx, 500, fmt.Sprintf("Failed to marshal request: %v", err))
	}
	
	// Send request to Python script
	responseJSON, status := r.vcPython.Process(string(requestJSON))
	if status != nil {
		return "", status
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		return "", log.ErrorNoErr(r.ctx, 500, fmt.Sprintf("Failed to parse response: %v", err))
	}
	
	if success, ok := response["success"].(bool); !ok || !success {
		errorMsg := "Unknown error"
		if errStr, ok := response["error"].(string); ok {
			errorMsg = errStr
		}
		return "", log.ErrorNoErr(r.ctx, 500, fmt.Sprintf("Speaker embedding extraction failed: %s", errorMsg))
	}
	
	embeddingPath, ok := response["embedding_path"].(string)
	if !ok || embeddingPath == "" {
		return "", log.ErrorNoErr(r.ctx, 500, "No embedding path in response")
	}
	
	return embeddingPath, nil
}

// EnsurePythonScriptReady initializes the Python subprocess for RVC
func (r *RVCAdapter) EnsurePythonScriptReady() *log.Status {
	if r.vcPython != nil {
		return nil // Already initialized
	}
	
	pythonPath := os.Getenv("FCBH_REVISE_AUDIO_VC_PYTHON")
	if pythonPath == "" {
		// Try to infer from conda environment
		if condaPrefix := os.Getenv("CONDA_PREFIX"); condaPrefix != "" {
			pythonPath = filepath.Join(condaPrefix, "bin", "python")
			if _, err := os.Stat(pythonPath); err != nil {
				pythonPath = ""
			}
		}
		if pythonPath == "" {
			return log.ErrorNoErr(r.ctx, 500, "FCBH_REVISE_AUDIO_VC_PYTHON environment variable not set")
		}
	}
	
	goproj := os.Getenv("GOPROJ")
	if goproj == "" {
		return log.ErrorNoErr(r.ctx, 500, "GOPROJ environment variable not set")
	}
	
	scriptPath := filepath.Join(goproj, "revise_audio", "python", "voice_conversion.py")
	if absPath, err := filepath.Abs(scriptPath); err == nil {
		scriptPath = absPath
	}
	
	var status *log.Status
	r.vcPython, status = stdio_exec.NewStdioExec(r.ctx, pythonPath, scriptPath)
	if status != nil {
		return status
	}
	
	return nil
}

// Close cleans up resources
func (r *RVCAdapter) Close() {
	if r.vcPython != nil {
		r.vcPython.Close()
		r.vcPython = nil
	}
}

