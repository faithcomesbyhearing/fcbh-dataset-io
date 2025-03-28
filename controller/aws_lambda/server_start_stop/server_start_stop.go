package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//{
//  "operation": "start", or "stop",
//   "instanceId": "i-0b22222aa0f43d1a5",
//   "bucket": "dataset-queue",
//   "folder": "input/"
//}

func handler(ctx context.Context, event events.CloudWatchEvent) error {
	var parm struct {
		Operation  string `json:"operation"`
		InstanceID string `json:"instance_id"`
		Bucket     string `json:"bucket"`
		Folder     string `json:"folder"`
	}
	var err error
	if err = json.Unmarshal(event.Detail, &parm); err != nil {
		return fmt.Errorf("error parsing detail: %v", err)
	}
	var cfg aws.Config
	cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion("us-west-2"))
	if err != nil {
		return fmt.Errorf("error loading AWS config: %v", err)
	}
	ec2Client := ec2.NewFromConfig(cfg)
	if parm.Operation == "start_asap" {
		err = startServer(ctx, ec2Client, parm.InstanceID)
	} else if parm.Operation == "stop_asap" {
		err = stopServer(ctx, ec2Client, parm.InstanceID)
	} else {
		var statusOutput *ec2.DescribeInstancesOutput
		statusOutput, err = ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			InstanceIds: []string{parm.InstanceID},
		})
		if err != nil {
			return fmt.Errorf("error describing instance: %v", err)
		}
		if len(statusOutput.Reservations) == 0 || len(statusOutput.Reservations[0].Instances) == 0 {
			return fmt.Errorf("instance %s not found", parm.InstanceID)
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
				Bucket:  aws.String(parm.Bucket),
				Prefix:  aws.String(parm.Folder),
				MaxKeys: aws.Int32(1),
			})
			if err != nil {
				return fmt.Errorf("error listing objects in queue: %v", err)
			}
			if parm.Operation == "start" && serverState == "stopped" && len(result.Contents) > 0 {
				err = startServer(ctx, ec2Client, parm.InstanceID)
			} else if parm.Operation == "stop" && serverState == "running" && len(result.Contents) == 0 {
				err = stopServer(ctx, ec2Client, parm.InstanceID)
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
