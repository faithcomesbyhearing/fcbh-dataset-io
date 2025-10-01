package update

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestInsertTimestampsFromEngnivn1daTimingsDB tests inserting timing data from engnivn1da_timings.db into init.db
func TestInsertTimestampsFromEngnivn1daTimingsDB(t *testing.T) {

	// Connect to engnivn1da_timings.db to read timing data
	timingsConn, err := sql.Open("sqlite3", "test_data/engnivn1da_timings.db")
	if err != nil {
		t.Fatalf("Failed to connect to engnivn1da_timings.db: %v", err)
	}
	defer timingsConn.Close()

	// Connect to init.db to insert timing data
	initConn, err := sql.Open("sqlite3", "test_data/init.db")
	if err != nil {
		t.Fatalf("Failed to connect to init.db: %v", err)
	}
	defer initConn.Close()

	// Get the file ID for John chapter 2 from init.db
	var fileId int64
	err = initConn.QueryRow("SELECT id FROM bible_files WHERE book_id = 'JHN' AND chapter_start = 2").Scan(&fileId)
	if err != nil {
		t.Fatalf("Failed to get file ID from init.db: %v", err)
	}

	// Read timing data from engnivn1da_timings.db for John chapter 2
	rows, err := timingsConn.Query(`
		SELECT verse_str, verse_end, script_begin_ts, script_end_ts 
		FROM scripts 
		WHERE book_id = 'JHN' AND chapter_num = 2 
		ORDER BY verse_num
	`)
	if err != nil {
		t.Fatalf("Failed to query engnivn1da_timings.db: %v", err)
	}
	defer rows.Close()

	var timestamps []Timestamp
	verseSeq := 0

	for rows.Next() {
		var verseStr, verseEnd string
		var beginTS, endTS float64

		err := rows.Scan(&verseStr, &verseEnd, &beginTS, &endTS)
		if err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}

		// Create timestamp record for insertion
		timestamp := Timestamp{
			TimestampId: 0, // 0 indicates new insert
			VerseStr:    verseStr,
			VerseEnd:    sql.NullString{String: verseEnd, Valid: verseEnd != ""},
			VerseSeq:    verseSeq,
			BeginTS:     beginTS,
			EndTS:       endTS,
		}

		timestamps = append(timestamps, timestamp)
		verseSeq++
	}

	if err = rows.Err(); err != nil {
		t.Fatalf("Error iterating rows: %v", err)
	}

	t.Logf("Found %d timestamps in engnivn1da_timings.db for John chapter 2", len(timestamps))

	// Clear existing timestamps for this file in init.db
	_, err = initConn.Exec("DELETE FROM bible_file_timestamps WHERE bible_file_id = ?", fileId)
	if err != nil {
		t.Fatalf("Failed to clear existing timestamps: %v", err)
	}

	// Insert new timestamps into init.db
	insertQuery := `INSERT INTO bible_file_timestamps 
		(bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) 
		VALUES (?, ?, ?, ?, ?, ?)`

	stmt, err := initConn.Prepare(insertQuery)
	if err != nil {
		t.Fatalf("Failed to prepare insert statement: %v", err)
	}
	defer stmt.Close()

	insertedCount := 0
	for _, ts := range timestamps {
		var verseEnd interface{}
		if ts.VerseEnd.Valid {
			verseEnd = ts.VerseEnd.String
		} else {
			verseEnd = nil
		}

		_, err = stmt.Exec(fileId, ts.VerseStr, verseEnd, ts.BeginTS, ts.EndTS, ts.VerseSeq)
		if err != nil {
			t.Fatalf("Failed to insert timestamp: %v", err)
		}
		insertedCount++
	}

	t.Logf("Inserted %d timestamps into init.db", insertedCount)

	// Verify the insertion
	var count int
	err = initConn.QueryRow("SELECT COUNT(*) FROM bible_file_timestamps WHERE bible_file_id = ?", fileId).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count inserted timestamps: %v", err)
	}

	if count != len(timestamps) {
		t.Errorf("Expected %d timestamps, got %d", len(timestamps), count)
	}

	// Verify specific timestamps
	var firstTimestamp Timestamp
	err = initConn.QueryRow(`
		SELECT id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence 
		FROM bible_file_timestamps 
		WHERE bible_file_id = ? AND verse_sequence = 0
	`, fileId).Scan(
		&firstTimestamp.TimestampId,
		&firstTimestamp.VerseStr,
		&firstTimestamp.VerseEnd,
		&firstTimestamp.BeginTS,
		&firstTimestamp.EndTS,
		&firstTimestamp.VerseSeq,
	)
	if err != nil {
		t.Fatalf("Failed to verify first timestamp: %v", err)
	}

	// Check that the first timestamp matches engnivn1da_timings.db data for John chapter 2
	if firstTimestamp.VerseStr != "0" {
		t.Errorf("Expected verse_str '0', got '%s'", firstTimestamp.VerseStr)
	}

	if firstTimestamp.BeginTS != 0.0 {
		t.Errorf("Expected begin_ts 0.0, got %f", firstTimestamp.BeginTS)
	}

	expectedEndTS := 2.35011407326921
	if abs(firstTimestamp.EndTS-expectedEndTS) > 0.0001 {
		t.Errorf("Expected end_ts %f, got %f", expectedEndTS, firstTimestamp.EndTS)
	}

	t.Logf("✅ Successfully inserted and verified %d timestamps from engnivn1da_timings.db into init.db", count)
}

// TestVerifyTimestampsMatch tests that the inserted timestamps match the source data
func TestVerifyTimestampsMatch(t *testing.T) {

	// Connect to both databases
	timingsConn, err := sql.Open("sqlite3", "test_data/engnivn1da_timings.db")
	if err != nil {
		t.Fatalf("Failed to connect to engnivn1da_timings.db: %v", err)
	}
	defer timingsConn.Close()

	initConn, err := sql.Open("sqlite3", "test_data/init.db")
	if err != nil {
		t.Fatalf("Failed to connect to init.db: %v", err)
	}
	defer initConn.Close()

	// Get file ID for John chapter 2
	var fileId int64
	err = initConn.QueryRow("SELECT id FROM bible_files WHERE book_id = 'JHN' AND chapter_start = 2").Scan(&fileId)
	if err != nil {
		t.Fatalf("Failed to get file ID: %v", err)
	}

	// Get timestamps from both databases
	timingsRows, err := timingsConn.Query(`
		SELECT verse_str, script_begin_ts, script_end_ts 
		FROM scripts 
		WHERE book_id = 'JHN' AND chapter_num = 2 
		ORDER BY verse_num
	`)
	if err != nil {
		t.Fatalf("Failed to query engnivn1da_timings.db: %v", err)
	}
	defer timingsRows.Close()

	initRows, err := initConn.Query(`
		SELECT verse_start, timestamp, timestamp_end 
		FROM bible_file_timestamps 
		WHERE bible_file_id = ? 
		ORDER BY verse_sequence
	`, fileId)
	if err != nil {
		t.Fatalf("Failed to query init.db: %v", err)
	}
	defer initRows.Close()

	// Compare timestamps
	timingsCount := 0
	initCount := 0

	for timingsRows.Next() {
		var timingsVerseStr string
		var timingsBeginTS, timingsEndTS float64

		err := timingsRows.Scan(&timingsVerseStr, &timingsBeginTS, &timingsEndTS)
		if err != nil {
			t.Fatalf("Failed to scan timings row: %v", err)
		}

		if !initRows.Next() {
			t.Errorf("init.db has fewer rows than engnivn1da_timings.db at position %d", timingsCount)
			break
		}

		var initVerseStr string
		var initBeginTS, initEndTS float64

		err = initRows.Scan(&initVerseStr, &initBeginTS, &initEndTS)
		if err != nil {
			t.Fatalf("Failed to scan init row: %v", err)
		}

		// Compare verse strings
		if timingsVerseStr != initVerseStr {
			t.Errorf("Verse mismatch at position %d: timings='%s', init='%s'", timingsCount, timingsVerseStr, initVerseStr)
		}

		// Compare timestamps (allow small floating point differences)
		if abs(timingsBeginTS-initBeginTS) > 0.0001 {
			t.Errorf("Begin timestamp mismatch at position %d: timings=%f, init=%f", timingsCount, timingsBeginTS, initBeginTS)
		}

		if abs(timingsEndTS-initEndTS) > 0.0001 {
			t.Errorf("End timestamp mismatch at position %d: timings=%f, init=%f", timingsCount, timingsEndTS, initEndTS)
		}

		timingsCount++
	}

	// Count remaining init rows
	for initRows.Next() {
		initCount++
	}

	if initCount > 0 {
		t.Errorf("init.db has %d more rows than engnivn1da_timings.db", initCount)
	}

	t.Logf("✅ Verified %d timestamps match between engnivn1da_timings.db and init.db", timingsCount)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
