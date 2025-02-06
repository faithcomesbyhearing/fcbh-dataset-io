package update

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetBoundaries(t *testing.T) {
	ctx := context.Background()
	conn, status := NewDBPAdapter(ctx)
	if status != nil {
		t.Fatal(status)
	}
	segments, status := conn.SelectTimestamps("ENGWEBN2DA", "MRK", 1)
	if status != nil {
		t.Fatal(status)
	}
	directory := filepath.Join(os.Getenv("FCBH_DATASET_FILES"), "ENGWEB", "ENGWEBN2DA")
	filename := filepath.Join(directory, "B02___01_Mark________ENGWEBN2DA.mp3")

	segments, status = GetBoundaries(ctx, filename, segments)
	if status != nil {
		t.Fatal(status)
	}
	for _, seg := range segments {
		fmt.Println(seg)
	}
	fmt.Println(len(segments))
}
