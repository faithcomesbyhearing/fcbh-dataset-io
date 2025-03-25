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
	// Create a buffer to build our message
	buf := new(bytes.Buffer)

	// Create a unique boundary for multipart message
	boundary := fmt.Sprintf("_boundary_%d", time.Now().Unix())

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
		fmt.Fprintf(buf, "%s: %s\r\n", key, value)
	}
	fmt.Fprintf(buf, "\r\n")

	// Text part
	fmt.Fprintf(buf, "--%s\r\n", boundary)
	fmt.Fprintf(buf, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(buf, "Content-Transfer-Encoding: 7bit\r\n\r\n")
	fmt.Fprintf(buf, "%s\r\n\r\n", bodyText)

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

		fmt.Fprintf(buf, "--%s\r\n", boundary)
		fmt.Fprintf(buf, "Content-Type: %s; name=\"%s\"\r\n", mimeType, fileName)
		fmt.Fprintf(buf, "Content-Disposition: attachment; filename=\"%s\"\r\n", fileName)
		fmt.Fprintf(buf, "Content-Transfer-Encoding: base64\r\n\r\n")

		// Encode attachment content to base64
		encoder := base64.NewEncoder(base64.StdEncoding, buf)
		encoder.Write(fileBytes)
		encoder.Close()
		fmt.Fprintf(buf, "\r\n")
	}

	// End multipart message
	fmt.Fprintf(buf, "--%s--\r\n", boundary)

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
