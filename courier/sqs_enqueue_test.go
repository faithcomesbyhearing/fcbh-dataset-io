package courier

import (
	"context"
	"fmt"
	"testing"
)

type Msg struct {
	Name    string
	Content string
}

func TestSQSEnqueue(t *testing.T) {
	ctx := context.Background()
	sqsURL := "https://sqs.us-west-2.amazonaws.com/078432969830/vessel_AP"
	var m = Msg{Name: "testName", Content: "testContent"}
	msgId, status := SQSEnqueue(ctx, sqsURL, m)
	if status != nil {
		t.Fatal(status)
	}
	fmt.Println("MessageId", msgId)
}

/*
aws sqs receive-message \
    --queue-url https://sqs.us-west-2.amazonaws.com/078432969830/vessel_AP \
    --max-number-of-messages 1

aws sqs delete-message \
   --queue-url https://sqs.us-west-2.amazonaws.com/078432969830/vessel_AP \
   --receipt-handle "AQEBwJnKyrHigUMZj6rYigCgxlaS3SLy0a..."
*/
