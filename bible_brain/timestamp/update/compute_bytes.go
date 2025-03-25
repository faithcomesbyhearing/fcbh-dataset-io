package update

import (
	"context"
	"encoding/json"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"math"
	"strconv"
)

type FramesResponse struct {
	Frames []Frame `json:"frames"`
}

type Frame struct {
	BestEffortTimestamp string `json:"best_effort_timestamp_time"`
	PacketPos           string `json:"pkt_pos"`
	PacketSize          string `json:"pkt_size"`
}

func ComputeBytes(ctx context.Context, file string, segments []Timestamp) ([]Timestamp, *log.Status) {
	if len(segments) == 0 {
		return segments, log.ErrorNoErr(ctx, 500, "no time segments provided")
	}
	probe, err := ffmpeg.Probe(file,
		ffmpeg.KwArgs{
			"show_frames":    "",
			"select_streams": "a",
			//"of":             "compact",
			"of":           "json",
			"show_entries": "frame=best_effort_timestamp_time,pkt_pos",
		})
	if err != nil {
		return segments, log.ErrorNoErr(ctx, 500, "probe error: %s", err.Error())
	}
	var response FramesResponse
	err = json.Unmarshal([]byte(probe), &response)
	if err != nil {
		return segments, log.ErrorNoErr(ctx, 500, "probe error: %s", err.Error())
	}
	var i int
	var time1, prevTime float64
	var pos, prevPos int64
	bound := segments[i].BeginTS
	for _, frame := range response.Frames {
		time1, err = strconv.ParseFloat(frame.BestEffortTimestamp, 64)
		if err != nil {
			log.Warn(ctx, "time parse error:", err.Error())
			continue
		}
		pos, err = strconv.ParseInt(frame.PacketPos, 10, 64)
		if err != nil {
			log.Warn(ctx, "position parse error:", err.Error())
			continue
		}
		if time1 >= bound {
			duration := time1 - prevTime
			nbytes := pos - prevPos
			segments[i].Duration = math.Round(duration*10000) / 10000
			segments[i].Position = prevPos
			segments[i].NumBytes = nbytes
			prevTime = time1
			prevPos = pos
			if i+1 != len(segments) {
				i++
				bound = segments[i].BeginTS
			} else {
				bound = 9999999.9 // search to end of pipe
			}
		}
	}
	return segments, nil
}
