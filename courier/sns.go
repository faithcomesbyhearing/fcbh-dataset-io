package courier

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

// SendMessage sends a text message to the specified phone number
func SendMessage(ctx context.Context, recipients []string, subject string, msg string) *log.Status {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-west-2"),
	)
	if err != nil {
		return log.Error(ctx, 500, err, "Failed to create SNS session to send message")
	}
	snsClient := sns.NewFromConfig(cfg)
	for _, phone := range recipients {
		var params sns.PublishInput
		params.Subject = aws.String(subject)
		params.Message = aws.String(msg)
		params.PhoneNumber = aws.String(phone)
		var result *sns.PublishOutput
		result, err = snsClient.Publish(ctx, &params)
		if err != nil {
			log.Warn(ctx, err, "Failed to send message to SNS")
		}
		log.Debug(ctx, "message Sent", *result.MessageId)
	}
	return nil
}
