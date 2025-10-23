package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestZipFileDirect(t *testing.T) {
	// Build paths using GOPROJ environment variable for portability
	goproj := os.Getenv("GOPROJ")
	if goproj == "" {
		t.Skip("GOPROJ environment variable not set - run 'source setup_env.sh' first")
	}

	audioZip := filepath.Join(goproj, "input/ENGWEB/ENGWEBN2DA/ENGWEBN2DA.zip")
	textFiles := filepath.Join(goproj, "input/ENGWEB/ENGWEBN_ET-usx/*.usx")

	zipFileYaml := fmt.Sprintf(`is_new: yes
dataset_name: 3a_zip_file
bible_id: ENGWEB
username: GaryNTest
output:
  sqlite: yes
audio_data:
  file: %s
text_data:
  file: %s
testament:
  nt_books: [MRK]
`, audioZip, textFiles)

	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 694})
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts WHERE script_begin_ts != 0.0", 0})
	DirectSqlTest(zipFileYaml, tests, t)
}
