package courier

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

func SQSEnqueue(ctx context.Context, queueURL string, data any) (string, *log.Status) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", log.Error(ctx, 500, err, "Error Marshalling SQS Message")
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-west-2"))
	if err != nil {
		return "", log.Error(ctx, 500, err, "Error loading AWS configuration")
	}
	client := sqs.NewFromConfig(cfg)
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(jsonData)),
	}
	result, err := client.SendMessage(ctx, input)
	if err != nil {
		return "", log.Error(ctx, 500, err, "Error Enqueueing SQS Message")
	}
	log.Info(ctx, "Enqueued: ", *result.MessageId)
	return *result.MessageId, nil
}
