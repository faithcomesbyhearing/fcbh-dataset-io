package courier

import (
	"context"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/go-gomail/gomail"
	"os"
	"strconv"
)

func GoMailSendMail(ctx context.Context, recipients []string, subject string, msg string,
	attachments []string) *log.Status {
	senderEmail := os.Getenv("SMTP_SENDER_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST_NAME")
	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_HOST_PORT"))

	m := gomail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", recipients...)
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
