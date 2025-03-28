package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"testing"
)

func TestStartHandler(t *testing.T) {
	ctx := context.Background()
	var event = events.CloudWatchEvent{}
	event.Detail = json.RawMessage(`{"operation": "start", 
		"instance_id": "i-0b22222aa0f43d1a5",
		"bucket": "dataset-queue",
		"folder": "input/"}`)
	err := handler(ctx, event)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
}

func TestStopHandler(t *testing.T) {
	ctx := context.Background()
	var event = events.CloudWatchEvent{}
	event.Detail = json.RawMessage(`{"operation": "stop", 
		"instance_id": "i-0b22222aa0f43d1a5",
		"bucket": "dataset-queue",
		"folder": "input/"}`)
	err := handler(ctx, event)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
}

func TestStartASAPHandler(t *testing.T) {
	ctx := context.Background()
	var event = events.CloudWatchEvent{}
	event.Detail = json.RawMessage(`{"operation": "start_asap", 
		"instance_id": "i-0b22222aa0f43d1a5",
		"bucket": "dataset-queue",
		"folder": "input/"}`)
	err := handler(ctx, event)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
}
