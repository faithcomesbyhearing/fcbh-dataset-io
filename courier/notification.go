package courier

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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
	//s3Client, status := b.preSignedURLClient()
	//if status == nil {
	//	for _, output := range b.outputs {
	//		message = append(message, output)
	//		key := b.createKey(b.run, "output", output)
	//		signedURL := b.genLongPreSignedURL(s3Client, key)
	//		message = append(message, signedURL)
	//	}
	return strings.Join(message, "\n\n")
}

func (b *Courier) preSignedURLClient() (*s3.S3, *log.Status) {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("us-west-2"),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, log.Error(b.ctx, 500, err, "unable to create S3 preSigned URL session")
	}
	return s3.New(sess), nil
}

func (b *Courier) genLongPreSignedURL(client *s3.S3, key string) string {
	// Important: Use v2 signing for longer expiration
	var input s3.GetObjectInput
	input.Bucket = aws.String(b.bucket)
	input.Key = aws.String(key)
	req, _ := client.GetObjectRequest(&input)
	// Set to use signature V2
	req.Config.S3ForcePathStyle = aws.Bool(true)
	req.Handlers.Sign.PushBack(func(r *request.Request) {
		r.ExpireTime = time.Hour * 24 * 30 // 30 days
	})
	// Generate the pre-signed URL
	url, err := req.Presign(30 * 24 * time.Hour)
	if err != nil {
		log.Warn(b.ctx, err, "unable to sign URL for", key)
	}
	return url
}
