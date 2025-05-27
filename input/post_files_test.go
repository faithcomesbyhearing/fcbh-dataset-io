package input

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestPostFiles(t *testing.T) {
	ctx := context.Background()
	post := NewPostFiles(ctx)
	defer post.RemoveDir()
	var filenames []string
	filename := filepath.Join(os.Getenv("FCBH_DATASET_FILES"), "ENGWEB", "ENGWEBN_ET.json")
	filenames = append(filenames, filename)
	for _, filename = range filenames {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatal(err)
		}
		post.ReadFile("text", file, filepath.Base(filename))
	}
	input, status := post.PostInput("text")
	if status != nil {
		t.Fatal(status)
	}
	if len(input) != 1 {
		t.Error("expected 1 input")
	} else {
		fmt.Println(input[0])
	}
}
