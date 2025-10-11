package update

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type HLSProcessor interface {
	ProcessFile(audioFile string, timestamps []Timestamp) (*HLSFileData, error)
}

type HLSFileData struct {
	File       HLSFile
	Bandwidths []HLSStreamBandwidth
	Bytes      []HLSStreamBytes
}

type LocalHLSProcessor struct {
	ctx      context.Context
	filesDir string
	bibleID  string
}

func NewLocalHLSProcessor(ctx context.Context, bibleID, timestampsFilesetID string) *LocalHLSProcessor {
	filesDir := os.Getenv("FCBH_DATASET_FILES")
	if filesDir == "" {
		filesDir = "/tmp/artie/files"
	}
	filesDir = filepath.Join(filesDir, bibleID, timestampsFilesetID)

	return &LocalHLSProcessor{
		ctx:      ctx,
		filesDir: filesDir,
		bibleID:  bibleID,
	}
}

func (p *LocalHLSProcessor) ProcessFile(audioFile string, timestamps []Timestamp) (*HLSFileData, error) {
	// Construct full path to audio file
	audioPath := filepath.Join(p.filesDir, audioFile)

	// Check if file exists
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Audio file not found: %s", audioPath)
	}

	// Get file info for size
	fileInfo, err := os.Stat(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}
	fileSize := fileInfo.Size()

	// Use FFmpeg to get audio bitrate
	bitrate, err := p.getAudioInfo(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio info: %v", err)
	}

	// Update bandwidth with actual bitrate
	actualBandwidth := bitrate
	if actualBandwidth == 0 {
		actualBandwidth = 64000 // fallback to 64kbps
	}

	// Get audio duration for SA filesets
	audioDuration, err := p.getAudioDuration(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio duration: %v", err)
	}

	// Validate that audio duration equals sum of all verse durations
	totalVerseDuration := 0.0
	for _, timestamp := range timestamps {
		totalVerseDuration += (timestamp.EndTS - timestamp.BeginTS)
	}

	// Allow for small floating point differences (within 1 second)
	if math.Abs(audioDuration-totalVerseDuration) > 1.0 {
		return nil, fmt.Errorf("audio duration mismatch: audio=%.2fs, sum of verses=%.2fs, difference=%.2fs",
			audioDuration, totalVerseDuration, math.Abs(audioDuration-totalVerseDuration))
	}

	// Create HLS file entry
	hlsFile := HLSFile{
		FileName:  replaceExtension(audioFile, ".m3u8"),
		FileSize:  fileSize,
		Duration:  int(audioDuration), // Round to int for SA filesets
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Create HLS stream bandwidth entry
	bandwidth := HLSStreamBandwidth{
		FileName:  replaceExtension(audioFile, "-64kbs.m3u8"),
		Bandwidth: actualBandwidth, // Use actual bitrate from FFmpeg
		Codec:     "mp4a.40.2",     // AAC codec
		Stream:    1,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Calculate byte offsets for each timestamp using FFmpeg (following AudioHLS.py approach)
	streamBytes, err := p.getBoundaries(audioPath, timestamps)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate byte boundaries: %v", err)
	}

	return &HLSFileData{
		File:       hlsFile,
		Bandwidths: []HLSStreamBandwidth{bandwidth},
		Bytes:      streamBytes,
	}, nil
}

// getAudioInfo uses FFmpeg to get audio bitrate (following AudioHLS.py approach)
func (p *LocalHLSProcessor) getAudioInfo(audioPath string) (bitrate int, err error) {
	// Use FFmpeg to get bitrate (following AudioHLS.py getBitrate method)
	cmd := exec.Command("ffprobe", "-select_streams", "a", "-v", "error", "-show_format", audioPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %v", err)
	}

	// Parse bitrate from output (following AudioHLS.py regex approach)
	outputStr := string(output)
	bitrateStart := strings.Index(outputStr, "bit_rate=")
	if bitrateStart == -1 {
		return 0, fmt.Errorf("bitrate not found in ffprobe output")
	}
	bitrateStart += len("bit_rate=")
	bitrateEnd := strings.Index(outputStr[bitrateStart:], "\n")
	if bitrateEnd == -1 {
		bitrateEnd = len(outputStr) - bitrateStart
	}
	bitrateStr := strings.TrimSpace(outputStr[bitrateStart : bitrateStart+bitrateEnd])
	bitrate, err = strconv.Atoi(bitrateStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse bitrate: %v", err)
	}

	return bitrate, nil
}

// getAudioDuration uses FFmpeg to get the actual audio duration
func (p *LocalHLSProcessor) getAudioDuration(audioPath string) (duration float64, err error) {
	// Use FFmpeg to get duration (following AudioHLS.py approach)
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", audioPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %v", err)
	}

	// Parse duration from output
	durationStr := strings.TrimSpace(string(output))
	duration, err = strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %v", err)
	}

	return duration, nil
}

// getBoundaries uses FFmpeg to get byte offsets for timestamps (following AudioHLS.py approach)
func (p *LocalHLSProcessor) getBoundaries(audioPath string, timestamps []Timestamp) ([]HLSStreamBytes, error) {
	// Get audio duration for validation
	audioDuration, err := p.getAudioDuration(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio duration: %v", err)
	}

	// Use FFmpeg to get frame data with timestamps and byte positions
	cmd := exec.Command("ffprobe", "-show_frames", "-select_streams", "a", "-of", "compact", "-show_entries", "frame=best_effort_timestamp_time,pkt_pos", audioPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %v", err)
	}

	// Parse the output to find byte positions for each timestamp
	lines := strings.Split(string(output), "\n")
	var streamBytes []HLSStreamBytes

	// Regex to match: frame|best_effort_timestamp_time=123.456|pkt_pos=789012
	timePosRegex := regexp.MustCompile(`best_effort_timestamp_time=([0-9.]+)\|pkt_pos=([0-9]+)`)

	var prevPos float64
	var currentTimestampIndex int

	// Process all frames to find boundaries
	for _, line := range lines {
		matches := timePosRegex.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}

		timestamp, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			continue
		}
		pos, err := strconv.ParseFloat(matches[2], 64)
		if err != nil {
			continue
		}

		// Check if we've reached the next timestamp boundary
		if currentTimestampIndex < len(timestamps) && timestamp >= timestamps[currentTimestampIndex].BeginTS {
			// Calculate bytes and offset for this segment
			var bytes, offset float64
			var timestampID int64

			if currentTimestampIndex == 0 {
				// First timestamp - segment from 0 to current position
				bytes = float64(pos - 0)
				offset = 0
			} else {
				// Subsequent timestamps - segment from previous to current position
				bytes = float64(pos - prevPos)
				offset = prevPos
			}

			// Runtime should be the duration of the corresponding timestamp
			runtime := timestamps[currentTimestampIndex].EndTS - timestamps[currentTimestampIndex].BeginTS
			timestampID = timestamps[currentTimestampIndex].TimestampId

			streamByte := HLSStreamBytes{
				Runtime:     runtime,
				Bytes:       int64(bytes),
				Offset:      int64(offset),
				TimestampID: timestampID,
				CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
				UpdatedAt:   time.Now().Format("2006-01-02 15:04:05"),
			}
			streamBytes = append(streamBytes, streamByte)

			prevPos = pos
			currentTimestampIndex++
		}
	}

	// Handle the last segment if we have remaining timestamps
	if currentTimestampIndex < len(timestamps) {

		// Use FFmpeg to get the byte position at the end of the file
		cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=size", "-of", "csv=p=0", audioPath)
		sizeOutput, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get file size: %v", err)
		}

		fileSize, err := strconv.ParseFloat(strings.TrimSpace(string(sizeOutput)), 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file size: %v", err)
		}

		// Runtime for the last segment should be the duration of the last timestamp
		runtime := timestamps[currentTimestampIndex].EndTS - timestamps[currentTimestampIndex].BeginTS
		bytes := int64(fileSize - prevPos)
		offset := int64(prevPos)

		streamByte := HLSStreamBytes{
			Runtime:     runtime,
			Bytes:       bytes,
			Offset:      offset,
			TimestampID: timestamps[currentTimestampIndex].TimestampId,
			CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt:   time.Now().Format("2006-01-02 15:04:05"),
		}
		streamBytes = append(streamBytes, streamByte)
	}

	// Validate that sum of all runtime values equals the audio duration
	totalRuntimeDuration := 0.0
	for _, streamByte := range streamBytes {
		totalRuntimeDuration += streamByte.Runtime
	}

	// Allow for small floating point differences (within 0.1 seconds)
	if math.Abs(audioDuration-totalRuntimeDuration) > 0.1 {
		return nil, fmt.Errorf("runtime duration mismatch: audio=%.2fs, sum of runtimes=%.2fs, difference=%.2fs",
			audioDuration, totalRuntimeDuration, math.Abs(audioDuration-totalRuntimeDuration))
	}

	return streamBytes, nil
}

func replaceExtension(filename, newExt string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return filename + newExt
	}
	return filename[:len(filename)-len(ext)] + newExt
}

// TODO: Implement LambdaHLSProcessor for future use
type LambdaHLSProcessor struct {
	ctx            context.Context
	lambdaFunction string
}

func NewLambdaHLSProcessor(ctx context.Context, lambdaFunction string) *LambdaHLSProcessor {
	return &LambdaHLSProcessor{
		ctx:            ctx,
		lambdaFunction: lambdaFunction,
	}
}

func (p *LambdaHLSProcessor) ProcessFile(audioFile string, timestamps []Timestamp) (*HLSFileData, error) {
	// TODO: Implement lambda call
	return nil, fmt.Errorf("lambda processor not implemented yet")
}
