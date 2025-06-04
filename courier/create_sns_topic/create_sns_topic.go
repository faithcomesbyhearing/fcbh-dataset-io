package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

const SNSTopic = "vessel_AP"

//  This is not used, delete if we deploy SQS instead

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}
	client := sns.NewFromConfig(cfg)
	input := &sns.CreateTopicInput{
		Name: aws.String(SNSTopic),
	}
	result, err := client.CreateTopic(ctx, input)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created topic: %s\n", *result.TopicArn)
}

// Created: arn:aws:sns:us-west-2:078432969830:vessel_AP
