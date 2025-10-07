package update

import (
	"context"
	"fmt"
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

	// Create HLS file entry
	hlsFile := HLSFile{
		FileName:  replaceExtension(audioFile, ".m3u8"),
		FileSize:  fileSize,
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

// getBoundaries uses FFmpeg to get byte offsets for timestamps (following AudioHLS.py approach)
func (p *LocalHLSProcessor) getBoundaries(audioPath string, timestamps []Timestamp) ([]HLSStreamBytes, error) {
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

	var prevTime, prevPos float64
	var currentTimestampIndex int

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
			// Calculate the segment for the previous timestamp
			if currentTimestampIndex > 0 {
				duration := timestamp - prevTime
				bytes := int64(pos - prevPos)
				offset := int64(prevPos)

				streamByte := HLSStreamBytes{
					Runtime:     duration,
					Bytes:       bytes,
					Offset:      offset,
					TimestampID: timestamps[currentTimestampIndex-1].TimestampId, // Use actual MySQL timestamp ID
					CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
					UpdatedAt:   time.Now().Format("2006-01-02 15:04:05"),
				}
				streamBytes = append(streamBytes, streamByte)
			} else {
				// Special case for the first timestamp (verse 0) - create a segment from 0 to current position
				duration := timestamp - 0.0
				bytes := int64(pos - 0)
				offset := int64(0)

				streamByte := HLSStreamBytes{
					Runtime:     duration,
					Bytes:       bytes,
					Offset:      offset,
					TimestampID: timestamps[currentTimestampIndex].TimestampId, // Use current timestamp ID
					CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
					UpdatedAt:   time.Now().Format("2006-01-02 15:04:05"),
				}
				streamBytes = append(streamBytes, streamByte)
			}

			prevTime = timestamp
			prevPos = pos
			currentTimestampIndex++
		}
	}

	// Handle the last segment if we have remaining timestamps
	if currentTimestampIndex < len(timestamps) {
		// For the last segment, we need to find the end of the file
		// This is a simplified approach - in production you'd want to be more precise
		fileInfo, err := os.Stat(audioPath)
		if err == nil {
			duration := float64(fileInfo.Size()) - prevPos
			bytes := int64(duration)
			offset := int64(prevPos)

			streamByte := HLSStreamBytes{
				Runtime:     duration,
				Bytes:       bytes,
				Offset:      offset,
				TimestampID: timestamps[currentTimestampIndex].TimestampId, // Use actual MySQL timestamp ID
				CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
				UpdatedAt:   time.Now().Format("2006-01-02 15:04:05"),
			}
			streamBytes = append(streamBytes, streamByte)
		}
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
