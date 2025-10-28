package timeout_notify

import (
	"context"
	"fmt"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"time"
)

// This is sample code provided by Claude.ai.
// It has not been implemented
// 10/27/25 GNG

type TimeoutNotify struct {
}

func (t *TimeoutNotify) RunJobWithTimeout(jobType string, job func(context.Context) error) error {
	// Define timeouts based on job category
	timeouts := map[string]time.Duration{
		"short":  5 * time.Minute,  // 3-4 min jobs → alert at 5 min
		"medium": 60 * time.Minute, // 45 min jobs → alert at 60 min
		"long":   8 * time.Hour,    // 4-6 hour jobs → alert at 8 hours
	}

	timeout := timeouts[jobType]
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Channel to track job completion
	done := make(chan error, 1)

	go func() {
		done <- job(ctx)
	}()

	// Monitor for timeout
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		// Job exceeded timeout - send alert!
		t.sendAlert(fmt.Sprintf("Job %s exceeded %v timeout", jobType, timeout))
		return fmt.Errorf("job timeout exceeded")
	}
}

func (s *TimeoutNotify) sendAlert(message string) {
	// Send email, Slack, SMS, etc.
	log.Error(context.Background(), 500, nil, message)
	// Could use AWS SNS, email, Slack webhook, etc.
}
