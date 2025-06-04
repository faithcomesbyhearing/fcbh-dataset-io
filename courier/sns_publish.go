package courier

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

// This is not used, delete if SQS solution is deployed

func PublishSNSMessage(ctx context.Context, topicArn string, subject string, event any) (string, *log.Status) {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return "", log.Error(ctx, 500, err, "Error Marshalling SNS Message")
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", log.Error(ctx, 500, err, "Error loading AWS configuration")
	}
	client := sns.NewFromConfig(cfg)
	input := &sns.PublishInput{
		Message:  aws.String(string(jsonData)),
		TopicArn: aws.String(topicArn),
		Subject:  aws.String(subject),
	}
	result, err := client.Publish(ctx, input)
	if err != nil {
		return "", log.Error(ctx, 500, err, "Error publishing SNS Message")
	}
	log.Info(ctx, "Published: ", subject, *result.MessageId)
	return *result.MessageId, nil
}
