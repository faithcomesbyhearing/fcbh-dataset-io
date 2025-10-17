package update

import (
	"context"
	"encoding/json"
	"math"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type FramesResponse struct {
	Frames []Frame `json:"frames"`
}

type Frame struct {
	BestEffortTimestamp string `json:"best_effort_timestamp_time"`
	PacketPos           string `json:"pkt_pos"`
	PacketSize          string `json:"pkt_size"`
}

// FrameData represents a single frame from ffprobe JSON output
type FrameData struct {
	Time float64
	Pos  int64
}

func ComputeBytes(ctx context.Context, file string, segments []Timestamp) ([]Timestamp, *log.Status) {
	if len(segments) == 0 {
		return segments, log.ErrorNoErr(ctx, 500, "no time segments provided")
	}
	probe, err := ffmpeg.Probe(file,
		ffmpeg.KwArgs{
			"show_frames":    "",
			"select_streams": "a",
			"of":             "json",
			"show_entries":   "frame=best_effort_timestamp_time,pkt_pos",
		})
	if err != nil {
		return segments, log.ErrorNoErr(ctx, 500, "probe error: %s", err.Error())
	}
	var response FramesResponse
	err = json.Unmarshal([]byte(probe), &response)
	if err != nil {
		return segments, log.ErrorNoErr(ctx, 500, "probe error: %s", err.Error())
	}

	// Parse all frames first
	frames, err := parseFramesFromJSON(response.Frames)
	if err != nil {
		return segments, log.ErrorNoErr(ctx, 500, "frame parse error: %s", err.Error())
	}

	if len(frames) == 0 {
		return segments, log.ErrorNoErr(ctx, 500, "no frames found in audio file")
	}

	// Find closest frame for each segment
	offsets := make([]int64, len(segments))
	for i, segment := range segments {
		closestFrame := findClosestFrameFromJSON(frames, segment.BeginTS)
		offsets[i] = closestFrame.Pos
	}

	// Get file size for last segment calculation
	fileSize, err := getFileSize(file)
	if err != nil {
		return segments, log.ErrorNoErr(ctx, 500, "file size error: %s", err.Error())
	}

	// Calculate bytes and update segments
	for i, segment := range segments {
		var bytes int64
		if i < len(segments)-1 {
			// bytes = offset[next] - offset[current]
			bytes = offsets[i+1] - offsets[i]
		} else {
			// Last segment: bytes = file_size - offset[current]
			bytes = fileSize - offsets[i]
		}

		// Calculate duration
		duration := segment.EndTS - segment.BeginTS
		segments[i].Duration = math.Round(duration*10000) / 10000
		segments[i].Position = offsets[i] // Cumulative offset from start of file
		segments[i].NumBytes = bytes
	}

	return segments, nil
}

// parseFramesFromJSON parses frames from ffmpeg JSON response
func parseFramesFromJSON(frames []Frame) ([]FrameData, error) {
	var result []FrameData
	for _, frame := range frames {
		time, err := strconv.ParseFloat(frame.BestEffortTimestamp, 64)
		if err != nil {
			continue
		}
		pos, err := strconv.ParseInt(frame.PacketPos, 10, 64)
		if err != nil {
			continue
		}
		result = append(result, FrameData{
			Time: time,
			Pos:  pos,
		})
	}
	return result, nil
}

// findClosestFrameFromJSON finds the frame with the minimum distance to the target timestamp
func findClosestFrameFromJSON(frames []FrameData, targetTime float64) FrameData {
	if len(frames) == 0 {
		return FrameData{}
	}

	closestFrame := frames[0]
	minDistance := math.Abs(frames[0].Time - targetTime)

	for _, frame := range frames[1:] {
		distance := math.Abs(frame.Time - targetTime)
		if distance < minDistance {
			minDistance = distance
			closestFrame = frame
		} else if distance > minDistance {
			// Frames are getting further away, we can break early
			break
		}
	}

	return closestFrame
}

// getFileSize gets the file size using ffprobe
func getFileSize(file string) (int64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=size", "-of", "csv=p=0", file)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	sizeStr := strings.TrimSpace(string(output))
	return strconv.ParseInt(sizeStr, 10, 64)
}
