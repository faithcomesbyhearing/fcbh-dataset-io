package main

/*
12factor note:
dataset_server provides an HTTP interface to the main service entry point
FIXME: move this to a lambda and deploy via serverless.yml

*/
import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"dataset/controller"
	"dataset/input"
	log "dataset/logger"
	//_ "net/http/pprof"
)

func main() {
	ctx := context.Background()
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/request", handler)
	log.Info(ctx, "Server starting on port 7777...")
	err := http.ListenAndServe(":7777", nil)
	if err != nil {
		log.Panic(ctx, "Error starting server: ", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := context.WithValue(context.Background(), `runType`, `server`)
	if r.Method != `POST` {
		errorResponse(ctx, w, http.StatusMethodNotAllowed, nil, `Only POST method is allowed`)
	}
	request, err := io.ReadAll(r.Body)
	if err != nil {
		errorResponse(ctx, w, http.StatusInternalServerError, err, `Error reading request to server`)
		return
	}
	control := controller.NewController(ctx, request)
	responder(ctx, w, control)
	log.Info(context.TODO(), time.Since(start))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := context.WithValue(context.Background(), `runType`, `server`)
	if r.Method != `POST` {
		errorResponse(ctx, w, http.StatusMethodNotAllowed, nil, `Only POST method is allowed`)
	}
	// Parse the multipart form with a max memory of 10 MiB
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		errorResponse(ctx, w, http.StatusInternalServerError, err, "Failed to parse multipart form")
		return
	}
	// Read form parts, currently tested to handle yaml file and one content file, either text or audio key
	postFiles := input.NewPostFiles(ctx)
	var request []byte
	var yamlHeader *multipart.FileHeader
	var dataHeader *multipart.FileHeader
	for key := range r.MultipartForm.File {
		file, header, err2 := r.FormFile(key)
		if err2 != nil {
			errorResponse(ctx, w, http.StatusInternalServerError, err2, "Failed to read multipart form")
			return
		}
		defer file.Close()
		if key == "yaml" {
			yamlHeader = header
			request, err = io.ReadAll(file)
			if err != nil {
				errorResponse(ctx, w, http.StatusInternalServerError, err, "Unable to read YAML file")
				return
			}
		} else {
			dataHeader = header
			status := postFiles.ReadFile(key, file, header.Filename)
			if status.IsErr {
				errorResponse(ctx, w, status.Status, nil, status.Message+status.Error())
			}
		}
	}
	log.Info(ctx, "Files uploaded successfully:", dataHeader.Filename, yamlHeader.Filename)
	control := controller.NewController(ctx, request)
	control.SetPostFiles(&postFiles)
	responder(ctx, w, control)
	log.Info(ctx, time.Since(start))
}

func responder(ctx context.Context, w http.ResponseWriter, control controller.Controller) {
	// 12factor note: this is a call to the main service entry point (control.ProcessV2)
	outputFiles, status := control.ProcessV2()
	if status.IsErr {
		w.WriteHeader(status.Status)
	} else {
		w.WriteHeader(200)
	}
	filename := outputFiles.FilePaths[0]
	var mimeType string
	if strings.HasSuffix(filename, `.csv`) {
		mimeType = "text/csv"
	} else if strings.HasSuffix(filename, `.json`) {
		mimeType = "application/json"
	} else {
		mimeType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimeType)
	var file *os.File
	file, err := os.Open(filename)
	if err != nil {
		errorResponse(ctx, w, http.StatusInternalServerError, err, `File containing results is not found`)
		return
	}
	defer file.Close()
	_, err = io.Copy(w, file)
	if err != nil {
		errorResponse(ctx, w, http.StatusInternalServerError, err, `Error writing results to http response`)
		return
	}
}

func errorResponse(ctx context.Context, w http.ResponseWriter, statusCode int, err error, message string) {
	status := log.Error(ctx, statusCode, err, message)
	http.Error(w, status.String(), statusCode)
}
