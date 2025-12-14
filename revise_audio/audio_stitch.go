package revise_audio

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/ffmpeg"
)

// AudioStitcher handles stitching audio segments together with cross-fades
type AudioStitcher struct {
	ctx     context.Context
	tempDir string
}

// NewAudioStitcher creates a new audio stitcher
func NewAudioStitcher(ctx context.Context) *AudioStitcher {
	return &AudioStitcher{
		ctx: ctx,
	}
}

// ReplaceSegmentInChapter replaces a segment in a chapter audio file with cross-fades
// originalFile: path to original chapter audio
// replacementFile: path to replacement audio snippet
// startTime: start time of segment to replace (seconds)
// endTime: end time of segment to replace (seconds)
// crossFadeDuration: duration of cross-fade at boundaries (seconds, default 0.05)
// Returns path to revised chapter audio file
func (s *AudioStitcher) ReplaceSegmentInChapter(
	originalFile string,
	replacementFile string,
	startTime float64,
	endTime float64,
	crossFadeDuration float64,
) (string, *log.Status) {
	// Allow zero cross-fade if explicitly requested (for testing)
	if crossFadeDuration < 0 {
		crossFadeDuration = 0.05 // Default 50ms cross-fade
	}
	if crossFadeDuration > 0.2 {
		crossFadeDuration = 0.2 // Max 200ms cross-fade
	}

	// Create temp directory if needed
	if s.tempDir == "" {
		var err error
		s.tempDir, err = os.MkdirTemp(os.Getenv("FCBH_DATASET_TMP"), "audio_stitch_")
		if err != nil {
			return "", log.Error(s.ctx, 500, err, "Error creating temp directory")
		}
	}

	// Get original audio duration using ffprobe
	originalDuration, status := s.getAudioDuration(originalFile)
	if status != nil {
		return "", status
	}

	// Calculate segment boundaries
	// If crossFadeDuration is 0, we still need to handle the boundaries correctly
	beforeEnd := startTime
	if crossFadeDuration > 0 {
		beforeEnd = startTime - crossFadeDuration
	}
	if beforeEnd < 0 {
		beforeEnd = 0
	}
	afterStart := endTime
	if crossFadeDuration > 0 {
		afterStart = endTime + crossFadeDuration
	}
	if afterStart > originalDuration {
		afterStart = originalDuration
	}

	// Extract segments
	beforeFile := filepath.Join(s.tempDir, fmt.Sprintf("before_%d.wav", time.Now().UnixNano()))
	afterFile := filepath.Join(s.tempDir, fmt.Sprintf("after_%d.wav", time.Now().UnixNano()))

	// Extract "before" segment (with fade-out at end if cross-fade > 0)
	if beforeEnd > 0 {
		beforeFileTmp, status := ffmpeg.ChopOneSegment(s.ctx, s.tempDir, originalFile, 0, beforeEnd)
		if status != nil {
			return "", status
		}
		// Apply fade-out to before segment only if cross-fade > 0
		if crossFadeDuration > 0 {
			beforeFile, status = s.applyFadeOut(beforeFileTmp, crossFadeDuration)
			if status != nil {
				return "", status
			}
		} else {
			beforeFile = beforeFileTmp
		}
	}

	// Apply cross-fades to replacement (fade-in at start, fade-out at end) only if cross-fade > 0
	var replacementWithFades string
	if crossFadeDuration > 0 {
		replacementWithFades, status = s.applyCrossFades(replacementFile, crossFadeDuration, crossFadeDuration)
		if status != nil {
			return "", status
		}
	} else {
		replacementWithFades = replacementFile
	}

	// Extract "after" segment (with fade-in at start if cross-fade > 0)
	if afterStart < originalDuration {
		afterFileTmp, status := ffmpeg.ChopOneSegment(s.ctx, s.tempDir, originalFile, afterStart, originalDuration)
		if status != nil {
			return "", status
		}
		// Apply fade-in to after segment only if cross-fade > 0
		if crossFadeDuration > 0 {
			afterFile, status = s.applyFadeIn(afterFileTmp, crossFadeDuration)
			if status != nil {
				return "", status
			}
		} else {
			afterFile = afterFileTmp
		}
	}

	// Concatenate segments
	outputFile := filepath.Join(s.tempDir, fmt.Sprintf("stitched_%d.wav", time.Now().UnixNano()))
	
	// Build list of files to concatenate
	var filesToConcat []string
	if beforeEnd > 0 {
		filesToConcat = append(filesToConcat, beforeFile)
	}
	filesToConcat = append(filesToConcat, replacementWithFades)
	if afterStart < originalDuration {
		filesToConcat = append(filesToConcat, afterFile)
	}

	// Concatenate using ffmpeg concat demuxer
	outputFile, status = s.concatAudioFiles(filesToConcat, outputFile)
	if status != nil {
		return "", status
	}

	return outputFile, nil
}

// getAudioDuration gets the duration of an audio file using ffprobe
func (s *AudioStitcher) getAudioDuration(audioFile string) (float64, *log.Status) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioFile)
	
	output, err := cmd.Output()
	if err != nil {
		return 0, log.Error(s.ctx, 500, err, "Error getting audio duration", audioFile)
	}

	var duration float64
	fmt.Sscanf(strings.TrimSpace(string(output)), "%f", &duration)
	if duration <= 0 {
		return 0, log.ErrorNoErr(s.ctx, 500, "Invalid audio duration", duration)
	}

	return duration, nil
}

// applyFadeIn applies a fade-in to an audio file
func (s *AudioStitcher) applyFadeIn(inputFile string, fadeDuration float64) (string, *log.Status) {
	outputFile := filepath.Join(s.tempDir, fmt.Sprintf("fadein_%d.wav", time.Now().UnixNano()))
	
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-af", fmt.Sprintf("afade=t=in:st=0:d=%.6f", fadeDuration),
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		"-y",
		outputFile)
	
	err := cmd.Run()
	if err != nil {
		return "", log.Error(s.ctx, 500, err, "Error applying fade-in")
	}
	
	return outputFile, nil
}

// applyFadeOut applies a fade-out to an audio file
func (s *AudioStitcher) applyFadeOut(inputFile string, fadeDuration float64) (string, *log.Status) {
	// Get duration first
	duration, status := s.getAudioDuration(inputFile)
	if status != nil {
		return "", status
	}
	
	outputFile := filepath.Join(s.tempDir, fmt.Sprintf("fadeout_%d.wav", time.Now().UnixNano()))
	fadeStart := duration - fadeDuration
	if fadeStart < 0 {
		fadeStart = 0
	}
	
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-af", fmt.Sprintf("afade=t=out:st=%.6f:d=%.6f", fadeStart, fadeDuration),
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		"-y",
		outputFile)
	
	err := cmd.Run()
	if err != nil {
		return "", log.Error(s.ctx, 500, err, "Error applying fade-out")
	}
	
	return outputFile, nil
}

// applyCrossFades applies fade-in and fade-out to an audio file
// Automatically adjusts fade duration for short snippets to preserve audible content
func (s *AudioStitcher) applyCrossFades(inputFile string, fadeInDuration float64, fadeOutDuration float64) (string, *log.Status) {
	// Get duration first
	duration, status := s.getAudioDuration(inputFile)
	if status != nil {
		return "", status
	}
	
	// For very short snippets, reduce fade duration to preserve audible content
	// Minimum audible duration after fades: 20ms
	minAudibleDuration := 0.020
	maxFadeDuration := (duration - minAudibleDuration) / 2.0
	if maxFadeDuration < 0 {
		maxFadeDuration = 0
	}
	
	// Adjust fade durations if needed
	actualFadeIn := fadeInDuration
	actualFadeOut := fadeOutDuration
	if fadeInDuration > maxFadeDuration {
		actualFadeIn = maxFadeDuration
	}
	if fadeOutDuration > maxFadeDuration {
		actualFadeOut = maxFadeDuration
	}
	
	// If snippet is too short for any fade, skip fades entirely
	if duration < minAudibleDuration {
		actualFadeIn = 0
		actualFadeOut = 0
	}
	
	outputFile := filepath.Join(s.tempDir, fmt.Sprintf("crossfade_%d.wav", time.Now().UnixNano()))
	
	// If no fades needed, just copy the file
	if actualFadeIn == 0 && actualFadeOut == 0 {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return "", log.Error(s.ctx, 500, err, "Error reading input file")
		}
		err = os.WriteFile(outputFile, data, 0644)
		if err != nil {
			return "", log.Error(s.ctx, 500, err, "Error writing output file")
		}
		return outputFile, nil
	}
	
	fadeOutStart := duration - actualFadeOut
	if fadeOutStart < 0 {
		fadeOutStart = 0
	}
	
	// Build filter: fade in at start, fade out at end (only if needed)
	var filterParts []string
	if actualFadeIn > 0 {
		filterParts = append(filterParts, fmt.Sprintf("afade=t=in:st=0:d=%.6f", actualFadeIn))
	}
	if actualFadeOut > 0 {
		filterParts = append(filterParts, fmt.Sprintf("afade=t=out:st=%.6f:d=%.6f", fadeOutStart, actualFadeOut))
	}
	
	filter := strings.Join(filterParts, ",")
	
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-af", filter,
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		"-y",
		outputFile)
	
	err := cmd.Run()
	if err != nil {
		return "", log.Error(s.ctx, 500, err, "Error applying cross-fades")
	}
	
	return outputFile, nil
}

// concatAudioFiles concatenates multiple audio files using ffmpeg concat demuxer
func (s *AudioStitcher) concatAudioFiles(inputFiles []string, outputFile string) (string, *log.Status) {
	if len(inputFiles) == 0 {
		return "", log.ErrorNoErr(s.ctx, 400, "No files to concatenate")
	}

	// Create concat file list
	concatFile := filepath.Join(s.tempDir, fmt.Sprintf("concat_%d.txt", time.Now().UnixNano()))
	file, err := os.Create(concatFile)
	if err != nil {
		return "", log.Error(s.ctx, 500, err, "Error creating concat file")
	}
	defer file.Close()

	for _, f := range inputFiles {
		// Use absolute path and escape single quotes
		absPath, err := filepath.Abs(f)
		if err != nil {
			return "", log.Error(s.ctx, 500, err, "Error getting absolute path", f)
		}
		// Format: file 'path'
		fmt.Fprintf(file, "file '%s'\n", strings.ReplaceAll(absPath, "'", "'\\''"))
	}
	file.Close()

	// Use concat demuxer
	cmd := exec.Command("ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		"-y",
		outputFile)

	err = cmd.Run()
	if err != nil {
		return "", log.Error(s.ctx, 500, err, "Error concatenating audio files")
	}

	return outputFile, nil
}

// Cleanup removes temporary files
func (s *AudioStitcher) Cleanup() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
		s.tempDir = ""
	}
}
