package courier

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"strings"
	"testing"
	"time"
)

func (b *Courier) Notification(status *log.Status, duration time.Duration) *log.Status {
	var st *log.Status
	if !testing.Testing() || b.IsUnitTest {
		var subject string
		var message string
		if status == nil {
			subject = "SUCCESS: " + b.dataset
			message = b.successMsg(duration)
		} else {
			subject = "FAILED: " + b.dataset
			message = b.failureMsg(status, duration)
		}
		st = SendMessage(b.ctx, recipients, subject, message)
		st = SendEmail(b.ctx, emailRecipients, subject, message)
	}
	return st
}

func (b *Courier) failureMsg(status *log.Status, duration time.Duration) string {
	var message []string
	message = append(message, "FAILED: "+b.dataset)
	message = append(message, status.Message)
	message = append(message, "Duration: "+duration.String())
	message = append(message, status.Trace)
	message = append(message, status.Request)
	return strings.Join(message, "\n")
}

func (b *Courier) successMsg(duration time.Duration) string {
	var message []string
	message = append(message, "SUCCESS: "+b.dataset)
	message = append(message, "Duration: "+duration.String())
	s3Client, status := b.presignedURLClient()
	if status == nil {
		for _, output := range b.outputs {
			message = append(message, output)
			key := b.createKey(b.run, "output", output)
			signedURL := b.genLongPresignedURL(s3Client, key)
			message = append(message, signedURL)
		}
	}
	return strings.Join(message, "\n")
}

func (b *Courier) presignedURLClient() (*s3.S3, *log.Status) {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("us-west-2"),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, log.Error(b.ctx, 500, err, "unable to create S3 presigned URL session")
	}
	return s3.New(sess), nil
}

func (b *Courier) genLongPresignedURL(client *s3.S3, key string) string {
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
