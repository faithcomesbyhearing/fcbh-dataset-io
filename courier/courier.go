package courier

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Courier struct {
	ctx         context.Context
	IsUnitTest  bool // Set to true by run_bucket_test.
	start       time.Time
	bucket      string
	username    string
	dataset     string
	run         int
	yamlContent string
	logFile     string
	databases   []string
	outputs     []string
	outputKeys  []string
}

func NewCourier(ctx context.Context, yaml []byte) Courier {
	var b Courier
	b.ctx = ctx
	b.start = time.Now()
	b.bucket = os.Getenv("FCBH_DATASET_IO_BUCKET")
	b.yamlContent = string(yaml)
	b.username = b.parseYaml(`username`)
	b.dataset = b.parseYaml(`dataset_name`)
	logFile := os.Getenv("FCBH_DATASET_LOG_FILE")
	if logFile != `` {
		b.AddLogFile(logFile)
	}
	return b
}

func (b *Courier) AddLogFile(logPath string) {
	b.logFile = logPath
	if !b.IsUnitTest {
		_ = os.Truncate(b.logFile, 0)
	}
}

func (b *Courier) AddDatabase(conn db.DBAdapter) {
	b.databases = append(b.databases, conn.DatabasePath)
}

func (b *Courier) AddOutput(outputPath string) {
	if len(outputPath) > 0 {
		b.outputs = append(b.outputs, outputPath)
	}
}

func (b *Courier) AddJson(records any, filePath string) {
	jsonData, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		log.Warn(b.ctx, err, "Failed to marshal ", filePath)
	} else {
		err = os.WriteFile(filePath, jsonData, 0644)
		if err != nil {
			log.Warn(b.ctx, err, "Failed to write ", filePath)
		} else {
			b.AddOutput(filePath)
		}
	}
}

func (b *Courier) GetOutputPaths() []string {
	return b.outputs
}

func (b *Courier) GetOutputByExt(fileExt string) []string {
	var results []string
	for _, path := range b.outputs {
		if strings.HasSuffix(path, fileExt) {
			results = append(results, path)
		}
	}
	return results
}

func (b *Courier) PersistToBucket() *log.Status {
	var allStatus []*log.Status
	var status *log.Status
	if !testing.Testing() || b.IsUnitTest {
		cfg, err := config.LoadDefaultConfig(b.ctx, config.WithRegion("us-west-2"))
		if err != nil {
			return log.Error(b.ctx, 500, err, "Error loading AWS config.")
		}
		client := s3.NewFromConfig(cfg)
		var run int
		run, status = b.findLastRun(client)
		allStatus = append(allStatus, status)
		run++
		_, status = b.uploadString(client, run, "request", b.dataset+".yaml", b.yamlContent)
		allStatus = append(allStatus, status)
		_, status = b.uploadFile(client, run, "log", b.logFile)
		allStatus = append(allStatus, status)
		for _, database := range b.databases {
			_, status = b.uploadFile(client, run, "database", database)
			allStatus = append(allStatus, status)
		}
		for _, output := range b.outputs {
			outputKey, status2 := b.uploadFile(client, run, "output", output)
			allStatus = append(allStatus, status2)
			b.outputKeys = append(b.outputKeys, outputKey)
		}
		loc, _ := time.LoadLocation("America/Denver")
		_, status = b.uploadString(client, run, "runtime", b.start.In(loc).Format(`Mon Jan 2 2006 03:04:05 pm MST`), "")
		allStatus = append(allStatus, status)
		_, status = b.uploadString(client, run, "duration", time.Since(b.start).String(), "")
		allStatus = append(allStatus, status)
		for _, stat := range allStatus {
			if stat != nil {
				status = stat
				break
			}
		}
	}
	return status
}

func (b *Courier) parseYaml(name string) string {
	var result string
	index := strings.Index(b.yamlContent, name+":")
	if index == -1 {
		result = "unknown-" + name
	} else {
		start := index + len(name) + 1
		end := strings.Index(b.yamlContent[start:], "\n")
		result = strings.TrimSpace(b.yamlContent[start : start+end])
	}
	return result
}

func (b *Courier) findLastRun(client *s3.Client) (int, *log.Status) {
	var result int
	var status *log.Status
	prefix := b.username + "/" + b.dataset + "/"
	output, err := client.ListObjectsV2(b.ctx, &s3.ListObjectsV2Input{
		Bucket: &b.bucket,
		Prefix: &prefix,
	})
	if err != nil {
		return result, log.Error(b.ctx, 500, err, "Error listing bucket objects.")
	}
	maxRun := 0
	for _, obj := range output.Contents {
		parts := strings.Split(*obj.Key, "/")
		if len(parts) < 4 {
			continue
		}
		runStr := parts[2]
		var runNum int
		runNum, err = strconv.Atoi(runStr)
		if err != nil {
			return result, log.Error(b.ctx, 500, err, "Error converting run number to int.")
		}
		if runNum > maxRun {
			maxRun = runNum
		}
	}
	return maxRun, status
}

func (b *Courier) uploadString(client *s3.Client, run int, typ string, filename string, content string) (string, *log.Status) {
	var objectKey string
	var status *log.Status
	objectKey = b.createKey(run, typ, filename)
	input := &s3.PutObjectInput{
		Bucket: &b.bucket,
		Key:    &objectKey,
		Body:   strings.NewReader(content),
	}
	_, err := client.PutObject(b.ctx, input)
	if err != nil {
		status = log.Error(b.ctx, 500, err, "Error uploading string content.")
	}
	return objectKey, status
}

func (b *Courier) uploadFile(client *s3.Client, run int, typ string, filePath string) (string, *log.Status) {
	var objectKey string
	var status *log.Status
	file, err := os.Open(filePath)
	if err != nil {
		log.Warn(b.ctx, 500, err, "Error opening file to upload to S3.")
		return objectKey, status
	}
	defer file.Close()
	objectKey = b.createKey(run, typ, filePath)
	_, err = client.PutObject(b.ctx, &s3.PutObjectInput{
		Bucket: &b.bucket,
		Key:    &objectKey,
		Body:   file,
	})
	if err != nil {
		status = log.Error(b.ctx, 500, err, "Error uploading file to S3.")
	}
	return objectKey, status
}

func (b *Courier) createKey(run int, typ string, filename string) string {
	runStr := fmt.Sprintf("%05d", run)
	filename = filepath.Base(filename)
	return b.username + "/" + b.dataset + "/" + runStr + "/" + typ + "/" + filename
}
