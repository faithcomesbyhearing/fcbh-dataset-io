package courier

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPerJobLogging(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir := t.TempDir()

	// Set environment variable for per-job logging
	os.Setenv("FCBH_DATASET_LOG_DIR", tmpDir)
	defer os.Unsetenv("FCBH_DATASET_LOG_DIR")

	// Create test YAML with username and dataset_name
	yaml := []byte(`
username: testuser
dataset_name: testdataset
`)

	// Create courier
	courier := NewCourier(context.Background(), yaml)
	courier.IsUnitTest = true

	// Verify log file was created in the directory
	if courier.logFile == "" {
		t.Fatal("Log file was not set")
	}

	// Verify log file is in the correct directory
	if !strings.HasPrefix(courier.logFile, tmpDir) {
		t.Errorf("Log file %s is not in expected directory %s", courier.logFile, tmpDir)
	}

	// Verify filename format: timestamp-username-datasetname.log
	basename := filepath.Base(courier.logFile)
	if !strings.Contains(basename, "-testuser-testdataset") {
		t.Errorf("Log filename %s does not match expected format timestamp-username-datasetname.log", basename)
	}
	if !strings.HasSuffix(basename, ".log") {
		t.Errorf("Log filename %s does not end with .log", basename)
	}
	// Verify timestamp is first (starts with digits)
	if len(basename) < 1 || (basename[0] < '0' || basename[0] > '9') {
		t.Errorf("Log filename %s should start with timestamp", basename)
	}

	// Verify log file doesn't exist yet (only created on first write)
	// but the directory should exist
	logDir := filepath.Dir(courier.logFile)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Errorf("Log directory %s was not created", logDir)
	}
}

func TestLegacyLogging(t *testing.T) {
	// Create temporary file for test log
	tmpFile := filepath.Join(t.TempDir(), "test.log")

	// Write some content to the file first
	if err := os.WriteFile(tmpFile, []byte("old content\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set environment variable for legacy logging
	os.Setenv("FCBH_DATASET_LOG_FILE", tmpFile)
	defer os.Unsetenv("FCBH_DATASET_LOG_FILE")

	// Make sure LOG_DIR is not set
	os.Unsetenv("FCBH_DATASET_LOG_DIR")

	// Create test YAML
	yaml := []byte(`
username: testuser
dataset_name: testdataset
`)

	// Create courier
	courier := NewCourier(context.Background(), yaml)
	courier.IsUnitTest = true

	// Manually call AddLogFile to test truncation
	courier.IsUnitTest = false
	courier.AddLogFile(tmpFile)

	// Verify file was truncated (should be 0 bytes)
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 0 {
		t.Errorf("Legacy log file was not truncated, size: %d", info.Size())
	}
}

func TestMultipleJobsSeparateLogs(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir := t.TempDir()

	os.Setenv("FCBH_DATASET_LOG_DIR", tmpDir)
	defer os.Unsetenv("FCBH_DATASET_LOG_DIR")

	yaml := []byte(`
username: testuser
dataset_name: testdataset
`)

	// Create first job
	courier1 := NewCourier(context.Background(), yaml)
	courier1.IsUnitTest = true
	logFile1 := courier1.logFile

	// Wait a moment to ensure different timestamp
	time.Sleep(time.Second * 2)

	// Create second job
	courier2 := NewCourier(context.Background(), yaml)
	courier2.IsUnitTest = true
	logFile2 := courier2.logFile

	// Verify they have different log files
	if logFile1 == logFile2 {
		t.Error("Two jobs created the same log file, expected different files")
	}

	// Verify both files are in the same directory
	if filepath.Dir(logFile1) != filepath.Dir(logFile2) {
		t.Error("Log files are in different directories")
	}
}

func TestNoLoggingEnvVar(t *testing.T) {
	// Make sure no logging env vars are set
	os.Unsetenv("FCBH_DATASET_LOG_DIR")
	os.Unsetenv("FCBH_DATASET_LOG_FILE")

	yaml := []byte(`
username: testuser
dataset_name: testdataset
`)

	// Create courier - should not panic
	courier := NewCourier(context.Background(), yaml)

	// Log file should be empty (logs to stderr instead)
	if courier.logFile != "" {
		t.Errorf("Expected no log file, got: %s", courier.logFile)
	}
}
