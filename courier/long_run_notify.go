package courier

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"strconv"
	"strings"
	"time"
)

type LongRunNotify struct {
	ctx         context.Context
	request     request.Request
	EstimateMin float64
	Threshold   time.Duration
}

func NewLongRunNotify(ctx context.Context, request request.Request) LongRunNotify {
	var l LongRunNotify
	l.ctx = ctx
	l.request = request
	var estimateMin float64 = 10.0
	if !request.Timestamps.NoTimestamps {
		estimateMin += 20.0
	}
	if !request.Training.NoTraining {
		estimateMin += 240.0
	}
	if !request.SpeechToText.NoSpeechToText {
		estimateMin += 20.0
	}
	if l.isVesselJob(request.NotifyOk) {
		estimateMin *= 0.3
	} else {
		estimateMin *= 2.0
	}
	log.Info(ctx, "Process will email if runs over", strconv.FormatFloat(estimateMin, 'g', 0, 64),
		"minutes.")
	l.EstimateMin = estimateMin
	l.Threshold = time.Duration(estimateMin*60.0) * time.Second
	return l
}

func (l LongRunNotify) SendEmail() {
	recipients := l.emails(l.request.NotifyErr)
	msg := "username: " + l.request.Username + "\n" +
		"dataset_name: " + l.request.DatasetName + "\n" +
		"Has been running for " + strconv.FormatFloat(l.EstimateMin, 'f', 1, 64) + " minutes."
	_ = GoMailSendMail(l.ctx, recipients, "Arti: Long Running Job", msg, nil)
}

func (l LongRunNotify) isVesselJob(addresses []string) bool {
	for _, a := range addresses {
		if strings.Contains(a, "sqs/") {
			return true
		}
	}
	return false
}

func (l LongRunNotify) emails(addresses []string) []string {
	var result []string
	for _, a := range addresses {
		if strings.Contains(a, "@") {
			result = append(result, a)
		}
	}
	return result
}
