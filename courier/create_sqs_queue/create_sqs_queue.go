package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"log"
)

const SQSQueueName = "vessel_AP"

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	client := sqs.NewFromConfig(cfg)
	input := &sqs.CreateQueueInput{
		QueueName: aws.String(SQSQueueName),
		Attributes: map[string]string{
			"DelaySeconds":           "0",
			"MessageRetentionPeriod": "1209600", // 14 days
		},
	}
	result, err := client.CreateQueue(ctx, input)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created queue: %s\n", *result.QueueUrl)
}

// Created queue: https://sqs.us-west-2.amazonaws.com/078432969830/vessel_AP
