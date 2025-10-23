package tests

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestZipFileDirect(t *testing.T) {
	// This test verifies ZIP file extraction by:
	// 1. Using audio files downloaded by previous tests (like TestLibrosaDirect)
	// 2. Creating a temporary zip file from those files
	// 3. Testing the system's ability to extract and process the zip

	fcbhFiles := os.Getenv("FCBH_DATASET_FILES")
	if fcbhFiles == "" {
		t.Skip("FCBH_DATASET_FILES is unset - please set and retest")
	}

	fcbhTmp := os.Getenv("FCBH_DATASET_TMP")
	if fcbhTmp == "" {
		t.Skip("FCBH_DATASET_TMP is unset - please set and retest")
	}

	// Check if Mark's audio files exist (downloaded by TestLibrosaDirect)
	audioDir := filepath.Join(fcbhFiles, "ENGWEB/ENGWEBN2DA")
	audioPattern := filepath.Join(audioDir, "B02*.mp3")
	audioFiles, _ := filepath.Glob(audioPattern)

	if len(audioFiles) == 0 {
		t.Skip("No ENGWEB audio files found - please run TestLibrosaDirect first")
	}

	// Create a temporary zip file from the downloaded audio files
	zipPath := filepath.Join(fcbhTmp, "ENGWEBN2DA-test.zip")
	err := createZipFromFiles(audioFiles, zipPath, "ENGWEBN2DA")
	if err != nil {
		t.Fatalf("Failed to create test zip file: %v", err)
	}
	defer os.Remove(zipPath) // Clean up the temporary zip

	// Use text files directly (no need to zip USX files for this test)
	textFiles := filepath.Join(fcbhFiles, "ENGWEB/ENGWEBN_ET-usx/*.usx")

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
`, zipPath, textFiles)

	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 694})
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts WHERE script_begin_ts != 0.0", 0})
	DirectSqlTest(zipFileYaml, tests, t)
}

// createZipFromFiles creates a zip archive from a list of files
func createZipFromFiles(files []string, zipPath string, baseDir string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		if err := addFileToZip(zipWriter, file, baseDir); err != nil {
			return err
		}
	}

	return nil
}

// addFileToZip adds a single file to a zip archive
func addFileToZip(zipWriter *zip.Writer, filePath string, baseDir string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file info for the header
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create zip header
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Preserve the filename structure inside the zip
	header.Name = filepath.Join(baseDir, filepath.Base(filePath))
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}
