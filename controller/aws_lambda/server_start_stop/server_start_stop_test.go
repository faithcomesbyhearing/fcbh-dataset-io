package main

import (
	"context"
	"testing"
)

func TestStopHandler(t *testing.T) {
	err := startStopTest("stop")
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
}

/*
func TestStartHandler(t *testing.T) {
	ctx := context.Background()
	var event = events.CloudWatchEvent{}
	event.Detail = json.RawMessage(`{"operation": "start",
		"instance_id": "i-0b22222aa0f43d1a5",
		"bucket": "dataset-queue",
		"folder": "input/"}`)
	err := handler(ctx, event.Detail)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
}
*/
/*
func TestStopHandler(t *testing.T) {
	ctx := context.Background()
	//var event = events.CloudWatchEvent{}
	////event := []byte(`{"operation": "stop",
	//	"instance_id": "i-0b22222aa0f43d1a5",
	//	"bucket": "dataset-queue",
	//	"folder": "input/"}`)
	var event = make(map[string]any)
	event["operation"] = "stop"
	event["instance_id"] = "i-0b22222aa0f43d1a5"
	event["bucket"] = "dataset-queue"
	event["folder"] = "input/"
	//event.Detail = json.RawMessage(`{"operation": "stop",
	//	"instance_id": "i-0b22222aa0f43d1a5",
	//	"bucket": "dataset-queue",
	//	"folder": "input/"}`)
	err := handler(ctx, event)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
}

*/
/*
func TestStartASAPHandler(t *testing.T) {
	ctx := context.Background()
	var event = events.CloudWatchEvent{}
	event.Detail = json.RawMessage(`{"operation": "start_asap",
		"instance_id": "i-0b22222aa0f43d1a5",
		"bucket": "dataset-queue",
		"folder": "input/"}`)
	err := handler(ctx, event.Detail)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}
}
*/

func startStopTest(operation string) error {
	ctx := context.Background()
	var event = make(map[string]any)
	event["operation"] = operation
	event["instance_id"] = "i-0b22222aa0f43d1a5"
	event["bucket"] = "dataset-queue"
	event["folder"] = "input/"
	err := handler(ctx, event)
	return err
}
