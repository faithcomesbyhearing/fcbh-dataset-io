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

// ProsodyMatcher interface for prosody matching operations
type ProsodyMatcher interface {
	// MatchProsody adjusts source audio to match reference audio's prosody
	// Returns path to adjusted audio file
	MatchProsody(sourceAudioPath string, referenceAudioPath string) (string, *log.Status)
	
	// ExtractProsodyFeatures extracts prosody features from audio
	// Returns prosody feature data (F0, energy, timing)
	ExtractProsodyFeatures(audioPath string) (*ProsodyFeatures, *log.Status)
}

// ProsodyFeatures represents extracted prosody information
type ProsodyFeatures struct {
	F0Mean      float64 // Mean fundamental frequency (Hz)
	F0Std       float64 // F0 standard deviation
	F0Contour   []float64 // F0 values over time
	EnergyMean  float64 // Mean energy (RMS)
	EnergyStd   float64 // Energy standard deviation
	SpeakingRate float64 // Words per second or similar metric
	Duration    float64 // Audio duration in seconds
}

// DSPProsodyAdapter implements ProsodyMatcher using DSP-based methods
type DSPProsodyAdapter struct {
	ctx      context.Context
	config   ProsodyConfig
	prosodyPython *stdio_exec.StdioExec
}

// NewDSPProsodyAdapter creates a new DSP-based prosody adapter
func NewDSPProsodyAdapter(ctx context.Context, config ProsodyConfig) *DSPProsodyAdapter {
	return &DSPProsodyAdapter{
		ctx:    ctx,
		config: config,
	}
}

// MatchProsody adjusts source audio to match reference audio's prosody
func (d *DSPProsodyAdapter) MatchProsody(sourceAudioPath string, referenceAudioPath string) (string, *log.Status) {
	if status := d.EnsurePythonScriptReady(); status != nil {
		return "", status
	}
	
	// Create temporary output file
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, fmt.Sprintf("prosody_output_%d.wav", os.Getpid()))
	
	// Prepare request
	request := map[string]interface{}{
		"mode":      "match",
		"source":    sourceAudioPath,
		"reference": referenceAudioPath,
		"output":    outputPath,
	}
	
	if d.config.SampleRate > 0 {
		request["sample_rate"] = d.config.SampleRate
	}
	if d.config.F0Method != "" {
		request["f0_method"] = d.config.F0Method
	}
	if d.config.PitchShiftRange > 0 {
		request["pitch_shift_range"] = d.config.PitchShiftRange
	}
	if d.config.TimeStretchRange > 0 {
		request["time_stretch_range"] = d.config.TimeStretchRange
	}
	
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", log.ErrorNoErr(d.ctx, 500, fmt.Sprintf("Failed to marshal request: %v", err))
	}
	
	// Send request to Python script
	responseJSON, status := d.prosodyPython.Process(string(requestJSON))
	if status != nil {
		return "", status
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		return "", log.ErrorNoErr(d.ctx, 500, fmt.Sprintf("Failed to parse response: %v", err))
	}
	
	if success, ok := response["success"].(bool); !ok || !success {
		errorMsg := "Unknown error"
		if errStr, ok := response["error"].(string); ok {
			errorMsg = errStr
		}
		return "", log.ErrorNoErr(d.ctx, 500, fmt.Sprintf("Prosody matching failed: %s", errorMsg))
	}
	
	outputPathStr, ok := response["output_path"].(string)
	if !ok || outputPathStr == "" {
		return "", log.ErrorNoErr(d.ctx, 500, "No output path in response")
	}
	
	return outputPathStr, nil
}

// ExtractProsodyFeatures extracts prosody features from audio
func (d *DSPProsodyAdapter) ExtractProsodyFeatures(audioPath string) (*ProsodyFeatures, *log.Status) {
	if status := d.EnsurePythonScriptReady(); status != nil {
		return nil, status
	}
	
	// Prepare request
	request := map[string]interface{}{
		"mode":   "extract",
		"source": audioPath,
	}
	
	if d.config.SampleRate > 0 {
		request["sample_rate"] = d.config.SampleRate
	}
	if d.config.F0Method != "" {
		request["f0_method"] = d.config.F0Method
	}
	
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, log.ErrorNoErr(d.ctx, 500, fmt.Sprintf("Failed to marshal request: %v", err))
	}
	
	// Send request to Python script
	responseJSON, status := d.prosodyPython.Process(string(requestJSON))
	if status != nil {
		return nil, status
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		return nil, log.ErrorNoErr(d.ctx, 500, fmt.Sprintf("Failed to parse response: %v", err))
	}
	
	if success, ok := response["success"].(bool); !ok || !success {
		errorMsg := "Unknown error"
		if errStr, ok := response["error"].(string); ok {
			errorMsg = errStr
		}
		return nil, log.ErrorNoErr(d.ctx, 500, fmt.Sprintf("Prosody feature extraction failed: %s", errorMsg))
	}
	
	featuresMap, ok := response["features"].(map[string]interface{})
	if !ok {
		return nil, log.ErrorNoErr(d.ctx, 500, "No features in response")
	}
	
	features := &ProsodyFeatures{}
	if f0Mean, ok := featuresMap["f0_mean"].(float64); ok {
		features.F0Mean = f0Mean
	}
	if f0Std, ok := featuresMap["f0_std"].(float64); ok {
		features.F0Std = f0Std
	}
	if energyMean, ok := featuresMap["energy_mean"].(float64); ok {
		features.EnergyMean = energyMean
	}
	if energyStd, ok := featuresMap["energy_std"].(float64); ok {
		features.EnergyStd = energyStd
	}
	if speakingRate, ok := featuresMap["speaking_rate"].(float64); ok {
		features.SpeakingRate = speakingRate
	}
	if duration, ok := featuresMap["duration"].(float64); ok {
		features.Duration = duration
	}
	
	return features, nil
}

// EnsurePythonScriptReady initializes the Python subprocess for prosody matching
func (d *DSPProsodyAdapter) EnsurePythonScriptReady() *log.Status {
	if d.prosodyPython != nil {
		return nil // Already initialized
	}
	
	pythonPath := os.Getenv("FCBH_REVISE_AUDIO_PROSODY_PYTHON")
	if pythonPath == "" {
		// Try to infer from conda environment
		if condaPrefix := os.Getenv("CONDA_PREFIX"); condaPrefix != "" {
			pythonPath = filepath.Join(condaPrefix, "bin", "python")
			if _, err := os.Stat(pythonPath); err != nil {
				pythonPath = ""
			}
		}
		if pythonPath == "" {
			return log.ErrorNoErr(d.ctx, 500, "FCBH_REVISE_AUDIO_PROSODY_PYTHON environment variable not set")
		}
	}
	
	goproj := os.Getenv("GOPROJ")
	if goproj == "" {
		return log.ErrorNoErr(d.ctx, 500, "GOPROJ environment variable not set")
	}
	
	scriptPath := filepath.Join(goproj, "revise_audio", "python", "prosody_match.py")
	if absPath, err := filepath.Abs(scriptPath); err == nil {
		scriptPath = absPath
	}
	
	var status *log.Status
	d.prosodyPython, status = stdio_exec.NewStdioExec(d.ctx, pythonPath, scriptPath)
	if status != nil {
		return status
	}
	
	return nil
}

// Close cleans up resources
func (d *DSPProsodyAdapter) Close() {
	if d.prosodyPython != nil {
		d.prosodyPython.Close()
		d.prosodyPython = nil
	}
}

