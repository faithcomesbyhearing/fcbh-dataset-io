package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

func sendEmailWithAttachments(recipients []string, subject string, bodyText string,
	attachments []string) error {
	sender := os.Getenv("SMTP_SENDER_EMAIL")
	boundary := fmt.Sprintf("_boundary_%d", time.Now().Unix())
	var buf bytes.Buffer
	buf.WriteString("From: " + sender + "\r\n")
	buf.WriteString("To: " + strings.Join(recipients, ", ") + "\r\n")
	buf.WriteString("Subject: " + subject + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n")
	buf.WriteString(`Content-Transfer-Encoding": "7bit"` + "\r\n")
	buf.WriteString("\r\n")
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	buf.WriteString(bodyText + "\r\n\r\n")
	for _, attachment := range attachments {
		fileBytes, err := os.ReadFile(attachment)
		if err != nil {
			return fmt.Errorf("failed to read attachment file %s: %v", attachment, err)
		}
		_, fileName := filepath.Split(attachment)
		var mimeType string
		switch strings.ToLower(filepath.Ext(fileName)) {
		case ".pdf":
			mimeType = "application/pdf"
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		case ".txt":
			mimeType = "text/plain"
		case ".csv":
			mimeType = "text/csv"
		case ".doc", ".docx":
			mimeType = "application/msword"
		case ".xls", ".xlsx":
			mimeType = "application/vnd.ms-excel"
		default:
			mimeType = "application/octet-stream"
		}
		buf.WriteString("--" + boundary + "\r\n")
		buf.WriteString("Content-Type: " + mimeType + "; name=\"" + fileName + "\"\r\n")
		buf.WriteString("Content-Disposition: attachment; filename=\"" + fileName + "\"\r\n")
		buf.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
		encoder := base64.NewEncoder(base64.StdEncoding, &buf)
		_, err = encoder.Write(fileBytes)
		if err != nil {
			return fmt.Errorf("failed to encode attachment %s: %v", fileName, err)
		}
		err = encoder.Close()
		if err != nil {
			return fmt.Errorf("failed to close encoder for %s: %v", fileName, err)
		}
		buf.WriteString("\r\n")
	}
	buf.WriteString("--" + boundary + "--\r\n")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	svc := ses.New(sess)
	input := &ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{
			Data: buf.Bytes(),
		},
	}
	_, err = svc.SendRawEmail(input)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}

func main() {
	err := sendEmailWithAttachments(
		[]string{"gary@shortsands.com", "gary.griswold@gmail.com"},
		"Email with Two Attachments",
		"This is the plain text body of the email.",
		[]string{"path/to/first_attachment.pdf", "path/to/second_attachment.csv"},
	)
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}
	fmt.Println("Email sent successfully!")
}
