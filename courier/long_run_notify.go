package courier

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"strconv"
	"strings"
	"time"
)

func LongRunNotify(ctx context.Context, request request.Request) {
	var estimateMin float64 = 9.6
	if !request.Timestamps.NoTimestamps {
		estimateMin += 20.0
	}
	if !request.Training.NoTraining {
		estimateMin += 240.0
	}
	if !request.SpeechToText.NoSpeechToText {
		estimateMin += 20.0
	}
	if isVesselJob(request.NotifyOk) {
		estimateMin *= 0.1
	} else {
		estimateMin *= 2.0
	}
	log.Info(ctx, "Process will email if runs over", strconv.FormatFloat(estimateMin, 'g', 0, 64),
		"minutes.")
	threshold := time.Duration(estimateMin*60.0) * time.Second

	done := make(chan struct{})
	go func() {
		select {
		case <-time.After(threshold):
			recipients := emails(request.NotifyErr)
			msg := "username: " + request.Username + "\n" +
				"dataset_name: " + request.DatasetName + "\n" +
				"Has been running for " + strconv.FormatFloat(estimateMin, 'f', 1, 64) + " minutes."
			_ = GoMailSendMail(ctx, recipients, "Arti: Long Running Job", msg, nil)
		case <-done:
			// Job completed before threshold - monitoring done
		}
	}()
}

func isVesselJob(addresses []string) bool {
	for _, a := range addresses {
		if strings.Contains(a, "sqs/") {
			return true
		}
	}
	return false
}

func emails(addresses []string) []string {
	var result []string
	for _, a := range addresses {
		if strings.Contains(a, "@") {
			result = append(result, a)
		}
	}
	return result
}
