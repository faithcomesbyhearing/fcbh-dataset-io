package revise_audio

import (
	"context"
	"encoding/json"
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
// preserveGaps: if true, measures original gap durations and trims replacement audio to match
// Returns path to revised chapter audio file
func (s *AudioStitcher) ReplaceSegmentInChapter(
	originalFile string,
	replacementFile string,
	startTime float64,
	endTime float64,
	crossFadeDuration float64,
) (string, *log.Status) {
	return s.ReplaceSegmentInChapterWithGapMatching(originalFile, replacementFile, startTime, endTime, crossFadeDuration, 0.0)
}

// ReplaceSegmentInChapterWithGapMatching replaces a segment using the step-by-step gap matching approach:
// 1. Measure leading silence in replacement audio
// 2. Measure original gap at boundary (or use provided originalGapBefore if > 0)
// 3. Place replacement such that total gap matches original:
//    - If replacement has no leading silence: place at end of original gap (no cross-fade)
//    - If replacement has leading silence: place with cross-fade to match original gap
func (s *AudioStitcher) ReplaceSegmentInChapterWithGapMatching(
	originalFile string,
	replacementFile string,
	boundaryTime float64, // Time where replacement starts (e.g., end of "called")
	endTime float64,      // End time of segment to replace
	crossFadeDuration float64, // Cross-fade duration (used only if replacement has leading silence)
	originalGapBefore float64, // If > 0, use this gap instead of measuring (allows precise calculation from word timestamps)
) (string, *log.Status) {
	// Create temp directory if needed
	if s.tempDir == "" {
		var err error
		s.tempDir, err = os.MkdirTemp(os.Getenv("FCBH_DATASET_TMP"), "audio_stitch_")
		if err != nil {
			return "", log.Error(s.ctx, 500, err, "Error creating temp directory")
		}
	}

	// Step 1: Measure leading silence in replacement audio
	fmt.Fprintf(os.Stderr, "[AudioStitcher] Measuring leading silence in replacement audio...\n")
	replacementLeadingSilence, status := s.measureLeadingSilence(replacementFile)
	if status != nil {
		replacementLeadingSilence = 0.0 // Default to 0 if measurement fails
	}
	fmt.Fprintf(os.Stderr, "[AudioStitcher] Replacement leading silence: %.3fs\n", replacementLeadingSilence)

	// Step 2: Measure original gap at boundary (gap between boundaryTime and next audio content)
	// If originalGapBefore is provided (> 0), use it; otherwise measure it
	if originalGapBefore <= 0 {
		fmt.Fprintf(os.Stderr, "[AudioStitcher] Measuring original gap at boundary...\n")
		measuredGapBefore, _, status := s.measureGapDurations(originalFile, boundaryTime, endTime)
		if status != nil {
			originalGapBefore = 0.0 // Default to 0 if measurement fails
		} else {
			originalGapBefore = measuredGapBefore
		}
	} else {
		fmt.Fprintf(os.Stderr, "[AudioStitcher] Using provided original gap at boundary: %.3fs\n", originalGapBefore)
	}
	fmt.Fprintf(os.Stderr, "[AudioStitcher] Original gap at boundary: %.3fs\n", originalGapBefore)

	// Step 3: Determine placement and cross-fade strategy
	// If replacement has no leading silence: place at end of original gap (no cross-fade)
	// If replacement has leading silence: place with cross-fade to match original gap
	useCrossFade := replacementLeadingSilence > 0.01 // Use cross-fade if > 10ms leading silence
	actualCrossFade := 0.0
	if useCrossFade {
		// Cross-fade duration should be the minimum of:
		// - Requested cross-fade duration
		// - Replacement leading silence (can't cross-fade more than available)
		// - Original gap (can't cross-fade more than available gap)
		actualCrossFade = crossFadeDuration
		if actualCrossFade > replacementLeadingSilence {
			actualCrossFade = replacementLeadingSilence
		}
		if actualCrossFade > originalGapBefore {
			actualCrossFade = originalGapBefore
		}
	}

	// Calculate where to place the replacement
	// If no cross-fade: place replacement at boundaryTime + originalGapBefore (end of gap)
	// If cross-fade: place replacement earlier to allow cross-fading the leading silence with the gap
	//   - We want to cross-fade replacementLeadingSilence with part of originalGapBefore
	//   - Place at: boundaryTime + (originalGapBefore - replacementLeadingSilence)
	//   - This ensures the total gap matches originalGapBefore after cross-fade
	replacementStartTime := boundaryTime + originalGapBefore
	if useCrossFade && replacementLeadingSilence > 0 {
		// Place earlier to allow cross-fade
		replacementStartTime = boundaryTime + (originalGapBefore - replacementLeadingSilence)
		// Ensure we don't go before boundary
		if replacementStartTime < boundaryTime {
			replacementStartTime = boundaryTime
		}
	}

	fmt.Fprintf(os.Stderr, "[AudioStitcher] Strategy: useCrossFade=%v, actualCrossFade=%.3fs, replacementStartTime=%.3fs\n",
		useCrossFade, actualCrossFade, replacementStartTime)

	// Get original audio duration
	originalDuration, status := s.getAudioDuration(originalFile)
	if status != nil {
		return "", status
	}

	// Extract "before" segment (up to replacement start)
	beforeFile := filepath.Join(s.tempDir, fmt.Sprintf("before_%d.wav", time.Now().UnixNano()))
	if replacementStartTime > 0 {
		beforeFileTmp, status := ffmpeg.ChopOneSegment(s.ctx, s.tempDir, originalFile, 0, replacementStartTime)
		if status != nil {
			return "", status
		}
		// Apply fade-out only if using cross-fade
		if useCrossFade && actualCrossFade > 0 {
			beforeFile, status = s.applyFadeOut(beforeFileTmp, actualCrossFade)
			if status != nil {
				return "", status
			}
		} else {
			beforeFile = beforeFileTmp
		}
	}

	// Process replacement audio
	// If using cross-fade, apply fade-in to replacement
	var replacementWithFades string
	if useCrossFade && actualCrossFade > 0 {
		replacementWithFades, status = s.applyFadeIn(replacementFile, actualCrossFade)
		if status != nil {
			return "", status
		}
	} else {
		replacementWithFades = replacementFile
	}

	// Extract "after" segment (from endTime onwards)
	afterFile := filepath.Join(s.tempDir, fmt.Sprintf("after_%d.wav", time.Now().UnixNano()))
	afterStart := endTime
	if afterStart < originalDuration {
		afterFileTmp, status := ffmpeg.ChopOneSegment(s.ctx, s.tempDir, originalFile, afterStart, originalDuration)
		if status != nil {
			return "", status
		}
		// Apply fade-in to after segment if using cross-fade
		if useCrossFade && actualCrossFade > 0 {
			afterFile, status = s.applyFadeIn(afterFileTmp, actualCrossFade)
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
	if replacementStartTime > 0 {
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

// ReplaceSegmentInChapterWithGapPreservation replaces a segment with optional gap preservation
// gapBefore: if > 0, use this gap duration instead of measuring (allows caller to specify from database)
// gapAfter: if > 0, use this gap duration instead of measuring
func (s *AudioStitcher) ReplaceSegmentInChapterWithGapPreservation(
	originalFile string,
	replacementFile string,
	startTime float64,
	endTime float64,
	crossFadeDuration float64,
	preserveGaps bool,
	gapBefore float64,
	gapAfter float64,
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

	// Measure original gap durations if preserving gaps
	// gapBefore: gap before startTime (between previous content and replacement start)
	// gapAfter: gap after endTime (between replacement end and next content)
	// If gapBefore/gapAfter are provided (> 0), use those instead of measuring
	if preserveGaps {
		if gapBefore <= 0 || gapAfter <= 0 {
			measuredGapBefore, measuredGapAfter, status := s.measureGapDurations(originalFile, startTime, endTime)
			if status != nil {
				// If measurement fails, continue without gap preservation
				preserveGaps = false
			} else {
				if gapBefore <= 0 {
					gapBefore = measuredGapBefore
				}
				if gapAfter <= 0 {
					gapAfter = measuredGapAfter
				}
			}
		}
		// gapBefore is the gap at the start boundary (what comes before the replacement)
		// This is what we need to preserve when stitching
	}

	// Trim replacement audio to match original gap durations
	var processedReplacement string
	if preserveGaps {
		processedReplacement, status = s.trimReplacementToMatchGaps(replacementFile, gapBefore, gapAfter)
		if status != nil {
			// If trimming fails, use original replacement
			processedReplacement = replacementFile
		}
	} else {
		processedReplacement = replacementFile
	}

	// Apply cross-fades to replacement (fade-in at start, fade-out at end) only if cross-fade > 0
	var replacementWithFades string
	if crossFadeDuration > 0 {
		replacementWithFades, status = s.applyCrossFades(processedReplacement, crossFadeDuration, crossFadeDuration)
		if status != nil {
			return "", status
		}
	} else {
		replacementWithFades = processedReplacement
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

// measureLeadingSilence measures the leading silence duration in an audio file
func (s *AudioStitcher) measureLeadingSilence(audioFile string) (float64, *log.Status) {
	// Use Python script to measure leading silence
	goproj := os.Getenv("GOPROJ")
	if goproj == "" {
		return 0.0, nil
	}
	
	pythonScript := filepath.Join(goproj, "revise_audio", "python", "measure_gaps.py")
	pythonPath := os.Getenv("FCBH_VITS_PYTHON")
	if pythonPath == "" {
		pythonPath = "python3"
	}
	
	// Get audio duration first
	duration, status := s.getAudioDuration(audioFile)
	if status != nil {
		return 0.0, nil
	}
	
	// Measure leading silence by checking from start of file
	// We'll pass start_time=0 and end_time=duration to measure leading gap
	cmd := exec.Command(pythonPath, pythonScript, audioFile,
		"0.0",
		fmt.Sprintf("%.6f", duration))
	
	output, err := cmd.Output()
	if err != nil {
		return 0.0, nil // Return 0 if measurement fails
	}
	
	// Parse JSON output
	var result struct {
		GapBefore float64 `json:"gap_before"`
		GapAfter  float64 `json:"gap_after"`
		Error     string  `json:"error,omitempty"`
	}
	
	if err := json.Unmarshal(output, &result); err != nil {
		return 0.0, nil
	}
	
	if result.Error != "" {
		return 0.0, nil
	}
	
	// GapBefore when measuring from start of file is the leading silence
	return result.GapBefore, nil
}

// measureGapDurations measures the silence gap durations before and after a segment in the original audio
func (s *AudioStitcher) measureGapDurations(originalFile string, startTime float64, endTime float64) (float64, float64, *log.Status) {
	// Use Python script to measure gaps
	goproj := os.Getenv("GOPROJ")
	if goproj == "" {
		return 0.0, 0.0, nil
	}
	
	pythonScript := filepath.Join(goproj, "revise_audio", "python", "measure_gaps.py")
	pythonPath := os.Getenv("FCBH_VITS_PYTHON")
	if pythonPath == "" {
		pythonPath = "python3"
	}
	
	cmd := exec.Command(pythonPath, pythonScript, originalFile,
		fmt.Sprintf("%.6f", startTime),
		fmt.Sprintf("%.6f", endTime))
	
	output, err := cmd.Output()
	if err != nil {
		return 0.0, 0.0, nil // Return 0 gaps if measurement fails
	}
	
	// Parse JSON output
	var result struct {
		GapBefore float64 `json:"gap_before"`
		GapAfter  float64 `json:"gap_after"`
		Error     string  `json:"error,omitempty"`
	}
	
	if err := json.Unmarshal(output, &result); err != nil {
		return 0.0, 0.0, nil
	}
	
	if result.Error != "" {
		return 0.0, 0.0, nil
	}
	
	return result.GapBefore, result.GapAfter, nil
}

// trimReplacementToMatchGaps trims trailing silence from replacement and adds back target gap
func (s *AudioStitcher) trimReplacementToMatchGaps(replacementFile string, targetGapBefore float64, targetGapAfter float64) (string, *log.Status) {
	// Use ffmpeg to trim silence, then pad with target gap amounts
	outputFile := filepath.Join(s.tempDir, fmt.Sprintf("trimmed_%d.wav", time.Now().UnixNano()))
	
	// First trim leading and trailing silence
	tempTrimmed := filepath.Join(s.tempDir, fmt.Sprintf("temp_trimmed_%d.wav", time.Now().UnixNano()))
	
	cmd := exec.Command("ffmpeg",
		"-i", replacementFile,
		"-af", "silenceremove=start_periods=1:start_duration=0.01:start_threshold=-40dB:detection=peak:stop_periods=1:stop_duration=0.01:stop_threshold=-40dB",
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		"-y",
		tempTrimmed)
	
	err := cmd.Run()
	if err != nil {
		return replacementFile, nil // Return original if trimming fails
	}
	
	// Add back target gaps using apad
	// First add leading silence if needed
	var currentFile = tempTrimmed
	if targetGapBefore > 0 {
		tempWithLeading := filepath.Join(s.tempDir, fmt.Sprintf("with_leading_%d.wav", time.Now().UnixNano()))
		cmd2 := exec.Command("ffmpeg",
			"-i", currentFile,
			"-af", fmt.Sprintf("apad=pad_dur=%.6f", targetGapBefore),
			"-acodec", "pcm_s16le",
			"-ar", "16000",
			"-ac", "1",
			"-y",
			tempWithLeading)
		
		if err := cmd2.Run(); err == nil {
			currentFile = tempWithLeading
		}
	}
	
	// Then add trailing silence if needed
	if targetGapAfter > 0 {
		cmd3 := exec.Command("ffmpeg",
			"-i", currentFile,
			"-af", fmt.Sprintf("apad=pad_dur=%.6f", targetGapAfter),
			"-acodec", "pcm_s16le",
			"-ar", "16000",
			"-ac", "1",
			"-y",
			outputFile)
		
		if err := cmd3.Run(); err != nil {
			return currentFile, nil
		}
		return outputFile, nil
	}
	
	// No trailing gap needed, just return current file
	if currentFile != tempTrimmed {
		return currentFile, nil
	}
	return tempTrimmed, nil
}

// Cleanup removes temporary files
func (s *AudioStitcher) Cleanup() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
		s.tempDir = ""
	}
}
