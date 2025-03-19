package courier

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

const emailSender = "apolyglot@fcbhmail.org"

// SendEmail sends an email to multiple recipients using AWS SES
func SendEmail(ctx context.Context, recipients []string, subject string, msg string) *log.Status {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	if err != nil {
		return log.Error(ctx, 500, err, "Error creating email session")
	}
	input := &ses.SendEmailInput{
		Source: aws.String(emailSender),
		Destination: &ses.Destination{
			ToAddresses: aws.StringSlice(recipients),
			//CcAddresses:  aws.StringSlice(email.Cc),
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Data:    aws.String(subject),
				Charset: aws.String("UTF-8"),
			},
			Body: &ses.Body{
				Text: &ses.Content{
					Data:    aws.String(msg),
					Charset: aws.String("UTF-8"),
				},
			},
		},
	}
	svc := ses.New(sess)
	result, err := svc.SendEmail(input)
	if err != nil {
		return log.Error(ctx, 500, err, "Error sending email")
	}
	log.Debug(ctx, "Email Sent", *result.MessageId)
	return nil
}
