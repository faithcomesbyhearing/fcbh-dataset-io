package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml"
	"gopkg.in/yaml.v3"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	cfg := GetConfig()
	yamlPath := GetArguments()
	yamlRequest, err := os.ReadFile(yamlPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ctx := context.Background()
	reqDecoder := decode_yaml.NewRequestDecoder(ctx)
	request, status := reqDecoder.Process(yamlRequest)
	if status != nil {
		fmt.Println(status)
		os.Exit(1)
	}
	var httpReq *http.Request
	if request.AudioData.POST != `` {
		httpReq = HttpMultiPost(cfg, yamlRequest, request.AudioData.POST, "audio")
	} else if request.TextData.POST != `` {
		httpReq = HttpMultiPost(cfg, yamlRequest, request.TextData.POST, "text")
	} else {
		httpReq = HttpPost(cfg, yamlRequest)
	}
	var fileType string
	if request.Output.CSV {
		fileType = ".csv"
	} else if request.Output.JSON {
		fileType = ".json"
	} else if request.Output.Sqlite {
		fileType = ".sqlite"
	} else {
		fileType = ".txt"
	}
	filename := filepath.Join(request.Output.Directory, request.DatasetName+fileType)
	statusCode := Response(filename, httpReq)
	DisplayOutput(filename)
	fmt.Println(statusCode, filename)
}

func GetArguments() string {
	if len(os.Args) < 2 {
		fmt.Println("Usage: FCBHDataset  filename.yaml")
		os.Exit(1)
	}
	return os.Args[1]
}

func HttpPost(cfg Config, request []byte) *http.Request {
	req, err := http.NewRequest("POST", cfg.Host+`/request`, bytes.NewBuffer(request))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/x-yaml")
	return req
}

func HttpMultiPost(cfg Config, yamlRequest []byte, filePath string, fType string) *http.Request {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	filePart, err := writer.CreateFormFile(fType, filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = io.Copy(filePart, file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	yamlPart, err := writer.CreateFormFile("yaml", "request.yaml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = io.Copy(yamlPart, bytes.NewBuffer(yamlRequest))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_ = writer.Close()
	req, err := http.NewRequest("POST", cfg.Host+`/upload`, body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func Response(outPath string, req *http.Request) string {
	var status string
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	status = resp.Status
	defer resp.Body.Close()
	file, err := os.Create(outPath)
	if err != nil {
		fmt.Println(err)
		return status
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println(err)
		return status
	}
	_ = file.Close()
	return status
}

func DisplayOutput(filename string) {
	if strings.HasSuffix(filename, ".sqlite") {
		return
	}
	if strings.HasSuffix(filename, ".db") {
		filename = strings.Replace(filename, ".db", ".sqlite", 1)
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		fmt.Println(scanner.Text())
		if lineCount == 20 {
			break
		}
	}
}

func Catch(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type Config struct {
	Host       string
	BBKey      string
	AWSProfile string
}

func GetConfig() Config {
	var cfg Config
	homeDir, err := os.UserHomeDir()
	Catch(err)
	var file *os.File
	cfgPath := filepath.Join(homeDir, "FCBHDataset.yaml")
	_, err = os.Stat(cfgPath)
	// Read Config
	if err == nil || !os.IsNotExist(err) {
		file, err = os.Open(cfgPath)
		Catch(err)
		decoder := yaml.NewDecoder(file)
		decoder.KnownFields(true)
		err = decoder.Decode(&cfg)
		Catch(err)
		err = file.Close()
		Catch(err)
	}
	isChanged := false
	if cfg.Host == `` {
		cfg.Host = Prompt(`Host Address`)
		isChanged = true
	}
	if cfg.BBKey == `` {
		cfg.BBKey = Prompt(`Bible Brain Key`)
		isChanged = true
	}
	if cfg.AWSProfile == `` {
		cfg.AWSProfile = Prompt(`AWS Profile`)
		isChanged = true
	}
	// Save Config
	if isChanged {
		bytes, err := yaml.Marshal(&cfg)
		Catch(err)
		file, err = os.OpenFile(cfgPath, os.O_WRONLY|os.O_CREATE, 0666)
		Catch(err)
		_, err = file.Write(bytes)
		Catch(err)
		_ = file.Close()

	}
	return cfg
}

func Prompt(prompt string) string {
	fmt.Print(`Enter `, prompt, ` : `)
	var answer string
	count, err := fmt.Scanln(&answer)
	if count > 0 {
		Catch(err)
	}
	return answer
}
