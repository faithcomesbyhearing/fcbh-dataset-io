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

// VITSAdapter handles So-VITS-SVC voice conversion
type VITSAdapter struct {
	ctx        context.Context
	config     VoiceConversionConfig
	vitsPython *stdio_exec.StdioExec
}

// NewVITSAdapter creates a new So-VITS-SVC adapter
func NewVITSAdapter(ctx context.Context, config VoiceConversionConfig) *VITSAdapter {
	return &VITSAdapter{
		ctx:    ctx,
		config: config,
	}
}

// ConvertVoice converts source audio to target speaker's voice using So-VITS-SVC
// modelPath is the path to the trained So-VITS-SVC model checkpoint (.pth file)
// configPath is the path to the So-VITS-SVC config.json file
// speaker is the speaker name/ID from the trained model
func (v *VITSAdapter) ConvertVoice(sourceAudioPath string, modelPath string, configPath string, speaker string) (string, *log.Status) {
	if status := v.EnsurePythonScriptReady(); status != nil {
		return "", status
	}
	
	// Create temporary output file
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, fmt.Sprintf("vits_output_%d.wav", os.Getpid()))
	
	// Prepare request
	request := map[string]interface{}{
		"mode":        "convert",
		"source":      sourceAudioPath,
		"model_path":  modelPath,
		"config_path": configPath,
		"output":      outputPath,
		"speaker":     speaker,
	}
	
	if v.config.F0Method != "" {
		request["f0_method"] = v.config.F0Method
	}
	if v.config.Device != "" {
		request["device"] = v.config.Device
	}
	
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", log.ErrorNoErr(v.ctx, 500, fmt.Sprintf("Failed to marshal request: %v", err))
	}
	
	// Send request to Python script
	responseJSON, status := v.vitsPython.Process(string(requestJSON))
	if status != nil {
		return "", status
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		return "", log.ErrorNoErr(v.ctx, 500, fmt.Sprintf("Failed to parse response: %v", err))
	}
	
	if success, ok := response["success"].(bool); !ok || !success {
		errorMsg := "Unknown error"
		if errStr, ok := response["error"].(string); ok {
			errorMsg = errStr
		}
		return "", log.ErrorNoErr(v.ctx, 500, fmt.Sprintf("So-VITS-SVC conversion failed: %s", errorMsg))
	}
	
	outputPathStr, ok := response["output_path"].(string)
	if !ok || outputPathStr == "" {
		return "", log.ErrorNoErr(v.ctx, 500, "No output path in response")
	}
	
	return outputPathStr, nil
}

// EnsurePythonScriptReady initializes the Python subprocess for So-VITS-SVC
func (v *VITSAdapter) EnsurePythonScriptReady() *log.Status {
	if v.vitsPython != nil {
		return nil // Already initialized
	}
	
	pythonPath := os.Getenv("FCBH_VITS_PYTHON")
	if pythonPath == "" {
		// Try to infer from conda environment
		if condaPrefix := os.Getenv("CONDA_PREFIX"); condaPrefix != "" {
			pythonPath = filepath.Join(condaPrefix, "bin", "python")
			if _, err := os.Stat(pythonPath); err != nil {
				pythonPath = ""
			}
		}
		if pythonPath == "" {
			return log.ErrorNoErr(v.ctx, 500, "FCBH_VITS_PYTHON environment variable not set")
		}
	}
	
	goproj := os.Getenv("GOPROJ")
	if goproj == "" {
		return log.ErrorNoErr(v.ctx, 500, "GOPROJ environment variable not set")
	}
	
	scriptPath := filepath.Join(goproj, "revise_audio", "vits", "python", "so_vits_svc_inference.py")
	if absPath, err := filepath.Abs(scriptPath); err == nil {
		scriptPath = absPath
	}
	
	var status *log.Status
	v.vitsPython, status = stdio_exec.NewStdioExec(v.ctx, pythonPath, scriptPath)
	if status != nil {
		return status
	}
	
	return nil
}

// Close cleans up resources
func (v *VITSAdapter) Close() {
	if v.vitsPython != nil {
		v.vitsPython.Close()
		v.vitsPython = nil
	}
}

