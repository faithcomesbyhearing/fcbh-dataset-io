package update

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestComputeBytes(t *testing.T) {
	ctx := context.Background()
	conn := getDBPConnection(t)
	defer conn.Close()
	hashId, status := conn.SelectHashId("ENGKJVN2DA")
	if status != nil {
		t.Fatal(status)
	}
	fileId, _, status := conn.SelectFileId(hashId, "MAT", 1)
	if status != nil {
		t.Fatal(status)
	}
	timestamps, status := conn.SelectTimestamps(fileId)
	if status != nil {
		t.Fatal(status)
	}
	directory := filepath.Join(os.Getenv("FCBH_DATASET_FILES"), "ENGKJV", "ENGKJVN2DA")
	filename := filepath.Join(directory, "B01___01_Matthew_____ENGKJVN2DA.mp3")
	fmt.Println(timestamps[0])
	timestamps, status = ComputeBytes(ctx, filename, timestamps)
	if status != nil {
		t.Fatal(status)
	}
	for _, seg := range timestamps {
		fmt.Println(seg)
	}
	fmt.Println(len(timestamps))
}
