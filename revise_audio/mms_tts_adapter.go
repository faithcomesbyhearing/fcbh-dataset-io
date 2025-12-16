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

// MMSTTSAdapter handles MMS TTS (Text-to-Speech) generation
type MMSTTSAdapter struct {
	ctx        context.Context
	lang       string
	mmsTtsPy   *stdio_exec.StdioExec
	tempDir    string
}

// NewMMSTTSAdapter creates a new MMS TTS adapter
func NewMMSTTSAdapter(ctx context.Context, lang string) (*MMSTTSAdapter, *log.Status) {
	adapter := &MMSTTSAdapter{
		ctx:  ctx,
		lang: lang,
	}
	
	// Create temp directory
	var err error
	adapter.tempDir, err = os.MkdirTemp(os.Getenv("FCBH_DATASET_TMP"), "mms_tts_")
	if err != nil {
		return nil, log.Error(ctx, 500, err, "Error creating temp directory for MMS TTS")
	}
	
	// Initialize Python script
	// Use GOPROJ to find the script (same pattern as other MMS modules)
	goproj := os.Getenv("GOPROJ")
	if goproj == "" {
		return nil, log.ErrorNoErr(ctx, 500, "GOPROJ environment variable not set")
	}
	pythonScript := filepath.Join(goproj, "revise_audio/python/mms_tts.py")
	
	// Verify script exists
	if _, err := os.Stat(pythonScript); os.IsNotExist(err) {
		return nil, log.Error(ctx, 500, err, "Python script not found", "path", pythonScript)
	}
	// Prefer MMS TTS-specific environment, fallback to MMS ASR, then system python
	pythonEnv := os.Getenv("FCBH_MMS_TTS_PYTHON")
	if pythonEnv == "" {
		pythonEnv = os.Getenv("FCBH_MMS_ASR_PYTHON")
	}
	if pythonEnv == "" {
		pythonEnv = "python3"
	}
	
	var status *log.Status
	adapter.mmsTtsPy, status = stdio_exec.NewStdioExec(ctx, pythonEnv, pythonScript, lang)
	if status != nil {
		return nil, status
	}
	
	return adapter, nil
}

// GenerateWord generates audio for a single word using MMS TTS
// Returns path to generated WAV file
func (a *MMSTTSAdapter) GenerateWord(word string) (string, *log.Status) {
	// Create output file path
	outputFile := filepath.Join(a.tempDir, fmt.Sprintf("tts_%s_%d.wav", word, os.Getpid()))
	
	// Prepare request
	request := map[string]interface{}{
		"text":        word,
		"output_path": outputFile,
		"seed":        555, // Deterministic generation
	}
	
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", log.Error(a.ctx, 500, err, "Error marshaling TTS request")
	}
	
	// Send request and get response
	responseJSON, status := a.mmsTtsPy.Process(string(requestJSON))
	if status != nil {
		return "", status
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		return "", log.Error(a.ctx, 500, err, "Error parsing TTS response", "response", responseJSON)
	}
	
	// Check for errors
	if errMsg, ok := response["error"].(string); ok {
		return "", log.ErrorNoErr(a.ctx, 500, fmt.Sprintf("TTS generation error: %s", errMsg))
	}
	
	// Get output path
	outputPath, ok := response["output_path"].(string)
	if !ok {
		return "", log.ErrorNoErr(a.ctx, 500, "TTS response missing output_path")
	}
	
	return outputPath, nil
}

// GeneratePhrase generates audio for a phrase (multiple words) using MMS TTS
// Returns path to generated WAV file
func (a *MMSTTSAdapter) GeneratePhrase(phrase string) (string, *log.Status) {
	// Create output file path
	outputFile := filepath.Join(a.tempDir, fmt.Sprintf("tts_phrase_%d_%d.wav", os.Getpid(), len(phrase)))
	
	// Prepare request
	request := map[string]interface{}{
		"text":        phrase,
		"output_path": outputFile,
		"seed":        555, // Deterministic generation
	}
	
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", log.Error(a.ctx, 500, err, "Error marshaling TTS request")
	}
	
	// Send request and get response
	responseJSON, status := a.mmsTtsPy.Process(string(requestJSON))
	if status != nil {
		return "", status
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		return "", log.Error(a.ctx, 500, err, "Error parsing TTS response", "response", responseJSON)
	}
	
	// Check for errors
	if errMsg, ok := response["error"].(string); ok {
		return "", log.ErrorNoErr(a.ctx, 500, fmt.Sprintf("TTS generation error: %s", errMsg))
	}
	
	// Get output path
	outputPath, ok := response["output_path"].(string)
	if !ok {
		return "", log.ErrorNoErr(a.ctx, 500, "TTS response missing output_path")
	}
	
	return outputPath, nil
}

// Close closes the adapter and cleans up resources
func (a *MMSTTSAdapter) Close() {
	if a.mmsTtsPy != nil {
		a.mmsTtsPy.Close()
	}
	if a.tempDir != "" {
		os.RemoveAll(a.tempDir)
	}
}

