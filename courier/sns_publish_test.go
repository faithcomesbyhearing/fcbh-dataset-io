package courier

import (
	"context"
	"fmt"
	"testing"
)

func TestPublishSNSMessage(t *testing.T) {
	ctx := context.Background()
	topicArn := "arn:aws:sns:us-west-2:078432969830:vessel_AP"
	subject := "test"
	type msg struct {
		name    string
		content string
	}
	var m = msg{name: "testName", content: "testContent"}
	msgId, status := PublishSNSMessage(ctx, topicArn, subject, m)
	if status != nil {
		t.Fatal(status)
	}
	fmt.Println("MessageId", msgId)
}
