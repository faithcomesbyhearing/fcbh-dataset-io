package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//{
//  "operation": "start", or "stop",
//   "instance_id": "i-0b22222aa0f43d1a5",
//   "bucket": "dataset-queue",
//   "folder": "input/"
//}

func handler(ctx context.Context, event map[string]any) error {
	fmt.Println("Starting AWS lambda handler", event)
	operation := event["operation"].(string)
	instanceId := event["instance_id"].(string)
	bucket := event["bucket"].(string)
	folder := event["folder"].(string)
	var err error
	var cfg aws.Config
	cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion("us-west-2"))
	if err != nil {
		return fmt.Errorf("error loading AWS config: %v", err)
	}
	ec2Client := ec2.NewFromConfig(cfg)
	if operation == "start_asap" {
		err = startServer(ctx, ec2Client, instanceId)
	} else if operation == "stop_asap" {
		err = stopServer(ctx, ec2Client, instanceId)
	} else {
		var statusOutput *ec2.DescribeInstancesOutput
		statusOutput, err = ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceId},
		})
		if err != nil {
			return fmt.Errorf("error describing instance: %v", err)
		}
		if len(statusOutput.Reservations) == 0 || len(statusOutput.Reservations[0].Instances) == 0 {
			return fmt.Errorf("instance %s not found", instanceId)
		}
		instance := statusOutput.Reservations[0].Instances[0]
		if instance.State == nil {
			return fmt.Errorf("instance state is nil")
		}
		serverState := instance.State.Name
		if serverState == "running" || serverState == "stopped" {
			s3Client := s3.NewFromConfig(cfg)
			var result *s3.ListObjectsV2Output
			result, err = s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
				Bucket:  aws.String(bucket),
				Prefix:  aws.String(folder),
				MaxKeys: aws.Int32(1),
			})
			if err != nil {
				return fmt.Errorf("error listing objects in queue: %v", err)
			}
			if operation == "start" && serverState == "stopped" && len(result.Contents) > 0 {
				err = startServer(ctx, ec2Client, instanceId)
			} else if operation == "stop" && serverState == "running" && len(result.Contents) == 0 {
				err = stopServer(ctx, ec2Client, instanceId)
			}
		}
	}
	return err
}

func startServer(ctx context.Context, client *ec2.Client, instanceId string) error {
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{instanceId},
	}
	_, err := client.StartInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("error starting instance: %v", err)
	}
	return nil
}

func stopServer(ctx context.Context, client *ec2.Client, instanceId string) error {
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{instanceId},
	}
	_, err := client.StopInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("error stopping instance %s: %v", instanceId, err)
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
