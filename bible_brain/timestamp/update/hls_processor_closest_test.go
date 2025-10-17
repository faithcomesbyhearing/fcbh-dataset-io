package update

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestHLSProcessorWithA19Data(t *testing.T) {
	// Test data paths
	testDataDir := "test_data"
	mp3File := testDataDir + "/A19__134_Psalms______ENGKJVO1DA-16k.mp3"
	timestampsFile := testDataDir + "/A19__134_Psalms______ENGKJVO1DA-16k_timestamps.csv"
	expectedBytesFile := testDataDir + "/A19__134_Psalms______ENGKJVO1DA-16k_bytes.csv"

	// Load timestamps with realistic durations
	timestamps, err := loadTimestampsFromCSVWithRealisticDurations(timestampsFile)
	if err != nil {
		t.Fatalf("Failed to load timestamps: %v", err)
	}

	// Load expected results
	expectedResults, err := loadExpectedBytesFromCSV(expectedBytesFile)
	if err != nil {
		t.Fatalf("Failed to load expected results: %v", err)
	}

	// Create HLS processor
	processor := &LocalHLSProcessor{}

	// Call getBoundaries
	result, err := processor.getBoundaries(mp3File, timestamps)
	if err != nil {
		t.Fatalf("getBoundaries failed: %v", err)
	}

	// Validate results
	if len(result) != len(expectedResults) {
		t.Fatalf("Expected %d segments, got %d", len(expectedResults), len(result))
	}

	for i, expected := range expectedResults {
		actual := result[i]

		t.Run("segment_"+string(rune(i)), func(t *testing.T) {
			if actual.Offset != expected.Offset {
				t.Errorf("Segment %d: expected offset %d, got %d", i, expected.Offset, actual.Offset)
			}
			if actual.Bytes != expected.Bytes {
				t.Errorf("Segment %d: expected bytes %d, got %d", i, expected.Bytes, actual.Bytes)
			}

			// Print actual vs expected for debugging
			t.Logf("Segment %d: offset=%d (expected %d), bytes=%d (expected %d)",
				i, actual.Offset, expected.Offset, actual.Bytes, expected.Bytes)
		})
	}
}

// loadTimestampsFromCSVWithRealisticDurations loads timestamps with realistic durations
func loadTimestampsFromCSVWithRealisticDurations(filename string) ([]Timestamp, error) {
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

// ExpectedResult represents the expected output from the bytes CSV
type ExpectedResult struct {
	VerseStart int
	Bytes      int64
	Offset     int64
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
