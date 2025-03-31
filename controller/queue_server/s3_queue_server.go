package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/cleanup"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/controller"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

const (
	inputFolder  = `input/`
	sucessFolder = `success/`
	failedFolder = `failed/`
)

func main() {
	ctx := context.WithValue(context.Background(), `runType`, `queue`)
	cleanup.CleanupDownloadDirectory(ctx)
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-west-2"),
		config.WithLogger(logger),
	)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err, "aws LoadDefaultConfig Failed In Queue Main")
		os.Exit(1)
	}
	client := s3.NewFromConfig(cfg)
	first := true
	for {
		bucketName := os.Getenv("FCBH_DATASET_QUEUE")
		fmt.Printf("FCBH_DATASET_QUEUE: %s", bucketName)

		object, key, status := getOldestObject(ctx, client, bucketName)
		if first && status != nil {
			_, _ = fmt.Fprintln(os.Stderr, err, "Reading First Input Failed In Queue Main")
			os.Exit(1)
		}
		first = false
		if status == nil && object != nil {
			control := controller.NewController(ctx, object)
			_, status = control.ProcessV2()
			var folder string
			if status != nil {
				folder = failedFolder
			} else {
				folder = sucessFolder
			}
			_ = moveOnCompletion(ctx, client, bucketName, key, folder)
		}
		time.Sleep(time.Second * 10)
	}
}

func getOldestObject(ctx context.Context, client *s3.Client, bucket string) ([]byte, string, *log.Status) {
	var content []byte
	var key string
	inFolder := inputFolder
	if runtime.GOOS == "darwin" {
		inFolder = "input_test/"
	}
	input := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: aws.String(inFolder),
	}
	result, err := client.ListObjectsV2(ctx, input)
	if err != nil {
		return content, key, log.Error(ctx, 500, err, "Error Listing Objects in Queue Input Folder")
	}
	if len(result.Contents) == 0 {
		return content, key, nil
	}
	sort.Slice(result.Contents, func(i, j int) bool {
		return result.Contents[i].LastModified.Before(*result.Contents[j].LastModified)
	})
	key = *result.Contents[0].Key
	getInput := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	object, err := client.GetObject(ctx, getInput)
	if err != nil {
		return content, key, log.Error(ctx, 500, err, "Error Getting Object in Queue Input Folder")
	}
	content, err = io.ReadAll(object.Body)
	_ = object.Body.Close()
	if err != nil {
		return content, key, log.Error(ctx, 500, err, "Error reading yaml file from Queue Input Folder.")
	}
	return content, key, nil
}

func moveOnCompletion(ctx context.Context, client *s3.Client, bucket, key string, folder string) *log.Status {
	source := bucket + "/" + key
	dateTime := time.Now().Local().Format("2006-01-02T15:04:05")
	target := folder + dateTime + "-" + strings.Split(key, "/")[1]
	_, err := client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &bucket,
		CopySource: &source,
		Key:        &target,
	})
	if err != nil {
		return log.Error(ctx, 500, err, "Error Moving File to", folder, "Folder")
	}
	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return log.Error(ctx, 500, err, "Error Deleting Object in Queue Input Folder")
	}
	return nil
}
