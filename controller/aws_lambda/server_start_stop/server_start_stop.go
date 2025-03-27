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
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//{
//  "operation": "start",
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
	if err := json.Unmarshal(event.Detail, &parm); err != nil {
		return fmt.Errorf("error parsing detail: %v", err)
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-west-2"))
	if err != nil {
		return fmt.Errorf("error loading AWS config: %v", err)
	}
	ec2Client := ec2.NewFromConfig(cfg)
	statusOutput, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
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
		result, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:  aws.String(parm.Bucket),
			Prefix:  aws.String(parm.Folder),
			MaxKeys: aws.Int32(1),
		})
		if err != nil {
			return fmt.Errorf("error listing objects in queue: %v", err)
		}
		if parm.Operation == "start" && serverState == "stopped" && len(result.Contents) > 0 {
			// load user_data.sh script
			userData := `#!/bin/bash
echo "Starting a_polyglot application $(date)" >> /var/log/app-startup.log
/home/ec2-user/go/src/fcbh-dataset-io/EC2_start_script.sh
`
			_, err = ec2Client.ModifyInstanceAttribute(ctx, &ec2.ModifyInstanceAttributeInput{
				InstanceId: aws.String(parm.InstanceID),
				UserData: &types.BlobAttributeValue{
					Value: []byte(userData),
				},
			})
			if err != nil {
				return fmt.Errorf("error updating user data: %v", err)
			}
			// Start the server
			_, err = ec2Client.StartInstances(ctx, &ec2.StartInstancesInput{
				InstanceIds: []string{parm.InstanceID},
			})
			if err != nil {
				return fmt.Errorf("error starting instance: %v", err)
			}
		} else if parm.Operation == "stop" && serverState == "running" && len(result.Contents) == 0 {
			input := &ec2.StopInstancesInput{
				InstanceIds: []string{parm.InstanceID},
			}
			_, err = ec2Client.StopInstances(ctx, input)
			if err != nil {
				return fmt.Errorf("error stopping instance %s: %v", parm.InstanceID, err)
			}
		}
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
