package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

func sendEmailWithAttachments(sender, recipient, subject, bodyText string, attachmentPaths []string) error {
	boundary := fmt.Sprintf("_boundary_%d", time.Now().Unix())
	var buf bytes.Buffer

	// Create a unique boundary for multipart message

	// Set up message headers
	headers := map[string]string{
		"From":                      sender,
		"To":                        recipient,
		"Subject":                   subject,
		"MIME-Version":              "1.0",
		"Content-Type":              fmt.Sprintf(`multipart/mixed; boundary="%s"`, boundary),
		"Content-Transfer-Encoding": "7bit",
	}

	// Write headers
	for key, value := range headers {
		buf.WriteString(key + ": " + value + "\r\n")
	}
	buf.WriteString("\r\n")

	// Text part
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	buf.WriteString(bodyText + "\r\n\r\n")

	// Add attachments
	for _, attachmentPath := range attachmentPaths {
		fileBytes, err := ioutil.ReadFile(attachmentPath)
		if err != nil {
			return fmt.Errorf("failed to read attachment file %s: %v", attachmentPath, err)
		}

		_, fileName := filepath.Split(attachmentPath)

		// Determine MIME type based on file extension
		mimeType := "application/octet-stream" // default
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
		}

		buf.WriteString("--" + boundary + "\r\n")
		buf.WriteString("Content-Type: " + mimeType + "; name=\"" + fileName + "\"\r\n")
		buf.WriteString("Content-Disposition: attachment; filename=\"" + fileName + "\"\r\n")
		buf.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

		// Encode attachment content to base64
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

	// End multipart message
	buf.WriteString("--" + boundary + "--\r\n")

	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"), // Replace with your AWS region
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	// Create an SES service client
	svc := ses.New(sess)

	// Send the email using the raw API
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
		"sender@example.com",
		"recipient@example.com",
		"Email with Two Attachments",
		"This is the plain text body of the email.",
		[]string{
			"path/to/first_attachment.pdf",
			"path/to/second_attachment.csv",
		},
	)
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}
	fmt.Println("Email sent successfully!")
}
