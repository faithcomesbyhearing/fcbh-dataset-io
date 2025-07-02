package courier

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"strings"
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
	var url = "https://sqs.us-west-2.amazonaws.com/078432969830/"
	var groupId = "default"
	parts := strings.Split(queueURL, "/")
	if len(parts) > 1 {
		url += parts[1]
	}
	if len(parts) > 2 {
		groupId = parts[2]
	}
	var hash = sha256.Sum256(jsonData)
	var deduplicationId = hex.EncodeToString(hash[:])
	input := &sqs.SendMessageInput{
		QueueUrl:               aws.String(url),
		MessageBody:            aws.String(string(jsonData)),
		MessageGroupId:         aws.String(groupId),
		MessageDeduplicationId: aws.String(deduplicationId),
	}
	result, err := client.SendMessage(ctx, input)
	if err != nil {
		return "", log.Error(ctx, 500, err, "Error Enqueueing SQS Message")
	}
	log.Info(ctx, "Enqueued: ", *result.MessageId)
	return *result.MessageId, nil
}
