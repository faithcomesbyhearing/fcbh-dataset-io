package courier

import (
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"strings"
	"testing"
	"time"
)

func (b *Courier) Notification(req request.Request, status *log.Status, duration time.Duration) *log.Status {
	//var st *log.Status
	if !testing.Testing() || b.IsUnitTest {
		var emailRecip []string
		var sqsURLS []string
		var subject string
		var message string
		var attachments []string
		if status == nil {
			emailRecip, sqsURLS = b.groupRecipients(req.NotifyOk)
			subject = "SUCCESS: " + b.dataset
			message = b.successMsg(duration)
			attachments = b.GetOutputByExt(".html")
		} else {
			emailRecip, sqsURLS = b.groupRecipients(req.NotifyErr)
			subject = "FAILED: " + b.dataset
			message = b.failureMsg(status, duration)
			attachments = append(attachments, b.logFile)
		}
		if len(emailRecip) > 0 {
			_ = GoMailSendMail(b.ctx, emailRecip, subject, message, attachments)
		}
		if len(sqsURLS) > 0 {
			jsonMessage := b.jsonMsg(duration, status == nil)
			for _, queueURL := range sqsURLS {
				_, status = SQSEnqueue(b.ctx, queueURL, jsonMessage)
				if status != nil {
					return status
				}
			}
		}
	}
	return nil // Do not propagate error
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
	for _, file := range b.outputKeys {
		if strings.HasSuffix(file, ".html") {
			message = append(message, "\t"+b.bucket+", "+file+"\n")
		}
	}
	return strings.Join(message, "\n\n")
}

type CompletionMsg struct {
	DatasetName string `yaml:"dataset_name"`
	Success     bool   `yaml:"success"`
	Completion  string `yaml:"completion"`
	Duration    string `yaml:"duration"`
	Bucket      string `yaml:"bucket"`
	Object      string `yaml:"object"`
}

func (b *Courier) jsonMsg(duration time.Duration, success bool) CompletionMsg {
	var msg CompletionMsg
	msg.DatasetName = b.dataset
	msg.Success = success
	denver, err := time.LoadLocation("America/Denver")
	if err != nil {
		msg.Completion = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	} else {
		msg.Completion = time.Now().In(denver).Format("2006-01-02T15:04:05Z")
	}
	msg.Duration = duration.Round(100 * time.Millisecond).String()
	msg.Bucket = b.bucket
	for _, file := range b.outputKeys {
		if strings.HasSuffix(file, "compare.json") {
			msg.Object = file
		}
	}
	return msg
}

func (b *Courier) groupRecipients(recipients []string) ([]string, []string) {
	var email []string
	var sqs []string
	for _, recip := range recipients {
		if strings.HasPrefix(recip, "sqs/") {
			url := strings.Replace(recip, "sqs/", "https://sqs.us-west-2.amazonaws.com/078432969830/", 1)
			sqs = append(sqs, url)
		} else {
			email = append(email, recip)
		}
	}
	return email, sqs
}

//sqs/vessel_AP
//https://sqs.us-west-2.amazonaws.com/078432969830/vessel_AP
