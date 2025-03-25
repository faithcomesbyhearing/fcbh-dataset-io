package courier

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/go-gomail/gomail"
	"os"
	"strconv"
	"strings"
)

func GoMailSendMail(ctx context.Context, recipients []string, subject string, msg string,
	attachments []string) *log.Status {
	senderEmail := os.Getenv("SMTP_SENDER_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST_NAME")
	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_HOST_PORT"))

	m := gomail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", strings.Join(recipients, ","))
	//m.SetAddressHeader("Cc", "cc1@company.com", "Dan")
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", msg)
	for _, file := range attachments {
		m.Attach(file)
	}
	d := gomail.NewDialer(smtpHost, smtpPort, senderEmail, password)
	err := d.DialAndSend(m)
	if err != nil {
		return log.Error(ctx, 500, err, "Error sending email")
	}
	log.Info(ctx, "Email sent", smtpHost, smtpPort, subject, recipients)
	return nil
}

// SESSendEmail sends an email to multiple recipients using AWS SES
func SESSendEmail(ctx context.Context, recipients []string, subject string, msg string,
	attachments []string) *log.Status {
	senderEmail := os.Getenv("SMTP_SENDER_EMAIL")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	if err != nil {
		return log.Error(ctx, 500, err, "Error creating email session")
	}
	input := &ses.SendEmailInput{
		Source: aws.String(senderEmail),
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
