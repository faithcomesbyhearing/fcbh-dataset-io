package ffmpeg

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type ProbeData struct {
	Format ProbeFormat `json:"format"`
}

type ProbeFormat struct {
	Filename       string `json:"filename"`
	NBStreams      int    `json:"nb_streams"`
	NBProgress     int    `json:"nb_programs"`
	NBStreamGroups int    `json:"nb_stream_groups"`
	FormatName     string `json:"format_name"`
	FormatLongName string `json:"format_long_name"`
	StartTime      string `json:"start_time"`
	Duration       string `json:"duration"`
	Size           string `json:"size"`
	BitRate        string `json:"bit_rate"`
	ProbeScore     int    `json:"probe_score"`
}

func GetAudioDuration(ctx context.Context, directory string, filename string) (float64, *log.Status) {
	var result float64
	probeData, status := GetProbeData(ctx, directory, filename)
	if status != nil {
		return result, status
	}
	if strings.TrimSpace(probeData.Format.Duration) != "" {
		var err error
		result, err = strconv.ParseFloat(probeData.Format.Duration, 64)
		if err != nil {
			return result, log.Error(ctx, 500, err, "Data conversion error in timestamp.GetAudioDuration")
		}
	} else {
		result, status = ComputeDuration(ctx, directory, filename)
		if status != nil {
			return result, status
		}
	}
	return result, nil
}

func GetAudioSize(ctx context.Context, directory string, filename string) (float64, *log.Status) {
	var result float64
	probeData, status := GetProbeData(ctx, directory, filename)
	if status != nil {
		return result, status
	}
	var err error
	result, err = strconv.ParseFloat(probeData.Format.Size, 64)
	if err != nil {
		return result, log.Error(ctx, 500, err, "Data conversion error in timestamp.GetAudioSize")
	}
	return result, nil
}

func GetAudioBitrate(ctx context.Context, directory string, filename string) (float64, *log.Status) {
	var result float64
	probeData, status := GetProbeData(ctx, directory, filename)
	if status != nil {
		return result, status
	}
	var err error
	result, err = strconv.ParseFloat(probeData.Format.BitRate, 64)
	if err != nil {
		return result, log.Error(ctx, 500, err, "Data conversion error in timestamp.GetAudioBitrate")
	}
	return result, nil
}

func GetProbeData(ctx context.Context, directory string, filename string) (ProbeData, *log.Status) {
	var result ProbeData
	filePath := filepath.Join(directory, filename)
	data, err := ffmpeg.Probe(filePath)
	if err != nil {
		return result, log.Error(ctx, 500, err, "Error in timestamp.GetProbeData")
	}
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		return result, log.Error(ctx, 500, err, "Error in timestamp.GetProbeData")
	}
	return result, nil
}
