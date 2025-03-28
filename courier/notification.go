package courier

import (
	req "github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func (b *Courier) Notification(req req.Request, status *log.Status, duration time.Duration) *log.Status {
	var st *log.Status
	if !testing.Testing() || b.IsUnitTest {
		var recipients []string
		var subject string
		var message string
		var attachments []string
		if status == nil {
			recipients = req.NotifyOk
			subject = "SUCCESS: " + b.dataset
			message = b.successMsg(duration)
			attachments = b.GetOutputByExt(".html")
		} else {
			recipients = req.NotifyErr
			subject = "FAILED: " + b.dataset
			message = b.failureMsg(status, duration)
		}
		st = GoMailSendMail(b.ctx, recipients, subject, message, attachments)
	}
	return st
}

func (b *Courier) failureMsg(status *log.Status, duration time.Duration) string {
	var message []string
	message = append(message, "FAILED: "+b.dataset)
	message = append(message, "Message: "+status.Message)
	message = append(message, "Duration: "+duration.Round(100*time.Millisecond).String())
	message = append(message, "Stack Trace: "+status.Trace)
	message = append(message, "Request: "+status.Request)
	return strings.Join(message, "\n\n")
}

func (b *Courier) successMsg(duration time.Duration) string {
	var message []string
	message = append(message, "SUCCESS: "+b.dataset)
	message = append(message, "Duration: "+duration.Round(100*time.Millisecond).String())
	message = append(message, "Output: ")
	files := b.GetOutputByExt(".html")
	for _, file := range files {
		message = append(message, "\t"+filepath.Base(file)+"\n")
	}
	return strings.Join(message, "\n\n")
}
