package update

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestComputeBytesWithA19Data(t *testing.T) {
	// Test data paths
	testDataDir := "test_data"
	mp3File := filepath.Join(testDataDir, "A19__134_Psalms______ENGKJVO1DA-16k.mp3")
	timestampsFile := filepath.Join(testDataDir, "A19__134_Psalms______ENGKJVO1DA-16k_timestamps.csv")
	expectedBytesFile := filepath.Join(testDataDir, "A19__134_Psalms______ENGKJVO1DA-16k_bytes.csv")

	// Check if test files exist
	if _, err := os.Stat(mp3File); os.IsNotExist(err) {
		t.Skipf("Test file %s does not exist", mp3File)
	}
	if _, err := os.Stat(timestampsFile); os.IsNotExist(err) {
		t.Skipf("Test file %s does not exist", timestampsFile)
	}
	if _, err := os.Stat(expectedBytesFile); os.IsNotExist(err) {
		t.Skipf("Test file %s does not exist", expectedBytesFile)
	}

	// Load timestamps
	timestamps, err := loadTimestampsFromCSV(timestampsFile)
	if err != nil {
		t.Fatalf("Failed to load timestamps: %v", err)
	}

	// Load expected results
	expectedResults, err := loadExpectedBytesFromCSV(expectedBytesFile)
	if err != nil {
		t.Fatalf("Failed to load expected results: %v", err)
	}

	// Call ComputeBytes
	ctx := context.Background()
	result, status := ComputeBytes(ctx, mp3File, timestamps)
	if status != nil {
		t.Fatalf("ComputeBytes failed: %v", status)
	}

	// Validate results
	if len(result) != len(expectedResults) {
		t.Fatalf("Expected %d segments, got %d", len(expectedResults), len(result))
	}

	for i, expected := range expectedResults {
		actual := result[i]

		t.Run(fmt.Sprintf("segment_%d", i), func(t *testing.T) {
			if actual.Position != expected.Offset {
				t.Errorf("Segment %d: expected offset %d, got %d", i, expected.Offset, actual.Position)
			}
			if actual.NumBytes != expected.Bytes {
				t.Errorf("Segment %d: expected bytes %d, got %d", i, expected.Bytes, actual.NumBytes)
			}

			// Print actual vs expected for debugging
			t.Logf("Segment %d: offset=%d (expected %d), bytes=%d (expected %d)",
				i, actual.Position, expected.Offset, actual.NumBytes, expected.Bytes)
		})
	}
}

// ExpectedResult represents the expected output from the bytes CSV
type ExpectedResult struct {
	VerseStart int
	Bytes      int64
	Offset     int64
}

// loadTimestampsFromCSV loads timestamps from the CSV file
func loadTimestampsFromCSV(filename string) ([]Timestamp, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var timestamps []Timestamp
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}
		if len(record) < 2 {
			continue
		}

		verseStart, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}
		timestamp, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			continue
		}

		// Create a timestamp with BeginTS and EndTS
		// Calculate realistic EndTS based on next timestamp or audio duration
		var endTS float64
		if i < len(records)-1 {
			// Use next timestamp as end time
			nextTimestamp, err := strconv.ParseFloat(records[i+1][1], 64)
			if err == nil {
				endTS = nextTimestamp
			} else {
				endTS = timestamp + 1.0 // Fallback
			}
		} else {
			// Last segment - use a reasonable duration
			endTS = timestamp + 5.0
		}

		ts := Timestamp{
			VerseStr: fmt.Sprintf("%d", verseStart),
			BeginTS:  timestamp,
			EndTS:    endTS,
			VerseSeq: verseStart,
		}
		timestamps = append(timestamps, ts)
	}

	return timestamps, nil
}

// loadExpectedBytesFromCSV loads expected results from the bytes CSV file
func loadExpectedBytesFromCSV(filename string) ([]ExpectedResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var results []ExpectedResult
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}
		if len(record) < 3 {
			continue
		}

		verseStart, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}
		bytes, err := strconv.ParseInt(record[1], 10, 64)
		if err != nil {
			continue
		}
		offset, err := strconv.ParseInt(record[2], 10, 64)
		if err != nil {
			continue
		}

		results = append(results, ExpectedResult{
			VerseStart: verseStart,
			Bytes:      bytes,
			Offset:     offset,
		})
	}

	return results, nil
}
