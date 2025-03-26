package cleanup

import (
	"context"
	"fmt"
	"testing"
)

func TestCleanupDirectory(t *testing.T) {
	ctx := context.Background()
	CleanupDownloadDirectory(ctx)
	fmt.Println("Successfully cleaned up old directories")
}
