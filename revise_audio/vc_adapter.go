package revise_audio

import (
	"context"
	"os"
	"path/filepath"
	
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/stdio_exec"
)

// VoiceConverter interface for voice conversion operations
type VoiceConverter interface {
	// ConvertVoice converts source audio to target speaker's voice
	// Returns path to converted audio file
	ConvertVoice(sourceAudioPath string, targetSpeakerEmbeddingPath string) (string, *log.Status)
	
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
func (r *RVCAdapter) ConvertVoice(sourceAudioPath string, targetSpeakerEmbeddingPath string) (string, *log.Status) {
	// TODO: Implement in task arti-lmh
	return "", log.ErrorNoErr(r.ctx, 500, "RVC voice conversion not yet implemented")
}

// ExtractSpeakerEmbedding extracts speaker embedding from audio
func (r *RVCAdapter) ExtractSpeakerEmbedding(audioPath string, outputPath string) (string, *log.Status) {
	// TODO: Implement in task arti-lmh
	return "", log.ErrorNoErr(r.ctx, 500, "Speaker embedding extraction not yet implemented")
}

// EnsurePythonScriptReady initializes the Python subprocess for RVC
func (r *RVCAdapter) EnsurePythonScriptReady() *log.Status {
	if r.vcPython != nil {
		return nil // Already initialized
	}
	
	pythonPath := os.Getenv("FCBH_REVISE_AUDIO_VC_PYTHON")
	if pythonPath == "" {
		return log.ErrorNoErr(r.ctx, 500, "FCBH_REVISE_AUDIO_VC_PYTHON environment variable not set")
	}
	
	scriptPath := filepath.Join(os.Getenv("GOPROJ"), "revise_audio/python/voice_conversion.py")
	
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

