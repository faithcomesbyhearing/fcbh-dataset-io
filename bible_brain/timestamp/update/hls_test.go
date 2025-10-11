package update

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

func TestHLSProcessor(t *testing.T) {
	// Set up test environment
	os.Setenv("FCBH_DATASET_DB", "/Users/jrstear/git/artie/bible_brain/timestamp/update/test_data/init.db")
	os.Setenv("FCBH_DATASET_FILES", "/Users/jrstear/git/artie/bible_brain/timestamp/update/test_data")

	ctx := context.Background()

	// Create test database connection
	conn := db.NewDBAdapter(ctx, "test_data/init.db")
	defer conn.Close()

	// Create HLS processor
	processor := NewLocalHLSProcessor(ctx, "ENGNIV", "ENGNIVN1DA")

	// Create test timestamps for John chapter 1
	testTimestamps := []Timestamp{
		{
			VerseStr:  "1",
			BeginTS:   0.0,
			EndTS:     2.5,
			AudioFile: "JHN_001.mp3",
		},
		{
			VerseStr:  "2",
			BeginTS:   2.5,
			EndTS:     5.0,
			AudioFile: "JHN_001.mp3",
		},
		{
			VerseStr:  "3",
			BeginTS:   5.0,
			EndTS:     7.5,
			AudioFile: "JHN_001.mp3",
		},
	}

	// Create a mock MP3 file for testing
	mockFile := "/Users/jrstear/git/artie/bible_brain/timestamp/update/test_data/ENGNIV/ENGNIVN1DA/JHN_001.mp3"
	os.MkdirAll("/Users/jrstear/git/artie/bible_brain/timestamp/update/test_data/ENGNIV/ENGNIVN1DA", 0755)
	file, err := os.Create(mockFile)
	if err != nil {
		t.Fatalf("Failed to create mock file: %v", err)
	}
	file.WriteString("mock mp3 content for testing")
	file.Close()
	defer os.Remove(mockFile)
	defer os.RemoveAll("/Users/jrstear/git/artie/bible_brain/timestamp/update/test_data/ENGNIV")

	// Process the file
	fileData, err := processor.ProcessFile("JHN_001.mp3", testTimestamps)
	if err != nil {
		t.Fatalf("Failed to process HLS file: %v", err)
	}

	// Verify the results
	if fileData.File.FileName != "JHN_001.m3u8" {
		t.Errorf("Expected file name 'JHN_001.m3u8', got '%s'", fileData.File.FileName)
	}

	if len(fileData.Bandwidths) != 1 {
		t.Errorf("Expected 1 bandwidth entry, got %d", len(fileData.Bandwidths))
	}

	if fileData.Bandwidths[0].Bandwidth != 64000 {
		t.Errorf("Expected bandwidth 64000, got %d", fileData.Bandwidths[0].Bandwidth)
	}

	if len(fileData.Bytes) != len(testTimestamps) {
		t.Errorf("Expected %d stream bytes entries, got %d", len(testTimestamps), len(fileData.Bytes))
	}

	// Verify stream bytes data
	for i, streamByte := range fileData.Bytes {
		expectedRuntime := testTimestamps[i].EndTS - testTimestamps[i].BeginTS
		if streamByte.Runtime != expectedRuntime {
			t.Errorf("Expected runtime %.2f, got %.2f", expectedRuntime, streamByte.Runtime)
		}

		if streamByte.Bytes <= 0 {
			t.Errorf("Expected positive bytes, got %d", streamByte.Bytes)
		}

		if streamByte.Offset < 0 {
			t.Errorf("Expected non-negative offset, got %d", streamByte.Offset)
		}
	}

	t.Logf("Successfully processed HLS file: %s", fileData.File.FileName)
	t.Logf("Generated %d stream bytes entries", len(fileData.Bytes))
}

func TestHLSProcessorWithRealData(t *testing.T) {
	// Set up test environment
	os.Setenv("FCBH_DATASET_DB", "/Users/jrstear/git/artie/bible_brain/timestamp/update/test_data/init.db")
	os.Setenv("FCBH_DATASET_FILES", "/Users/jrstear/tmp/artie/files")
	os.Setenv("DBP_MYSQL_DSN", "root:@tcp(localhost)/jrs")

	ctx := context.Background()

	// Create test database connection
	conn := db.NewDBAdapter(ctx, "test_data/init.db")
	defer conn.Close()

	// Create DBP adapter for database operations
	dbpConn, status := NewDBPAdapter(ctx)
	if status != nil {
		t.Fatalf("Failed to create DBP adapter: %v", status)
	}
	defer dbpConn.Close()

	// Create HLS processor
	processor := NewLocalHLSProcessor(ctx, "ENGNIV", "ENGNIVN1DA")

	// Process both John chapters 1 and 2
	chapters := []int{1, 2}
	totalFiles := 0
	totalTimestamps := 0

	for _, chapter := range chapters {
		// Get timestamps for this chapter from bible_file_timestamps table
		query := `SELECT bft.id, bft.verse_start, bft.timestamp, bft.timestamp_end, bf.file_name 
				  FROM bible_file_timestamps bft 
				  JOIN bible_files bf ON bft.bible_file_id = bf.id 
				  WHERE bf.book_id = ? AND bf.chapter_start = ? 
				  ORDER BY bft.verse_sequence`

		rows, err := conn.DB.Query(query, "JHN", chapter)
		if err != nil {
			t.Fatalf("Failed to query timestamps for JHN chapter %d: %v", chapter, err)
		}
		defer rows.Close()

		var timestamps []Timestamp
		verseSeq := 1
		for rows.Next() {
			var id int64
			var verseStart string
			var beginTS, endTS float64
			var fileName string

			err := rows.Scan(&id, &verseStart, &beginTS, &endTS, &fileName)
			if err != nil {
				t.Fatalf("Failed to scan timestamp for JHN chapter %d: %v", chapter, err)
			}

			timestamps = append(timestamps, Timestamp{
				TimestampId: id,
				VerseStr:    verseStart,
				BeginTS:     beginTS,
				EndTS:       endTS,
				AudioFile:   fileName,
				VerseSeq:    verseSeq,
			})
			verseSeq++
		}

		if len(timestamps) == 0 {
			t.Logf("No timestamps found for JHN chapter %d, skipping", chapter)
			continue
		}

		// Get the audio file name for this chapter
		audioFile := fmt.Sprintf("B04___%02d_John________ENGNIVN1DA.mp3", chapter)

		// Process the file
		fileData, err := processor.ProcessFile(audioFile, timestamps)
		if err != nil {
			t.Fatalf("Failed to process HLS file for chapter %d: %v", chapter, err)
		}

		// Verify the results
		expectedFileName := audioFile[:len(audioFile)-4] + ".m3u8" // Replace .mp3 with .m3u8
		if fileData.File.FileName != expectedFileName {
			t.Errorf("Chapter %d: Expected file name '%s', got '%s'", chapter, expectedFileName, fileData.File.FileName)
		}

		if len(fileData.Bandwidths) != 1 {
			t.Errorf("Chapter %d: Expected 1 bandwidth entry, got %d", chapter, len(fileData.Bandwidths))
		}

		if fileData.Bandwidths[0].Bandwidth != 64000 {
			t.Errorf("Chapter %d: Expected bandwidth 64000, got %d", chapter, fileData.Bandwidths[0].Bandwidth)
		}

		if len(fileData.Bytes) != len(timestamps) {
			t.Errorf("Chapter %d: Expected %d stream bytes entries, got %d", chapter, len(timestamps), len(fileData.Bytes))
		}

		// Verify stream bytes data
		for i, streamByte := range fileData.Bytes {
			// Runtime should equal the duration of the corresponding timestamp
			expectedRuntime := timestamps[i].EndTS - timestamps[i].BeginTS

			// Allow for small floating point differences (within 0.1 seconds)
			if math.Abs(streamByte.Runtime-expectedRuntime) > 0.1 {
				t.Errorf("Chapter %d, timestamp %d: Expected runtime %.2f, got %.2f", chapter, i, expectedRuntime, streamByte.Runtime)
			}

			if streamByte.Bytes <= 0 {
				t.Errorf("Chapter %d, timestamp %d: Expected positive bytes, got %d", chapter, i, streamByte.Bytes)
			}

			if streamByte.Offset < 0 {
				t.Errorf("Chapter %d, timestamp %d: Expected non-negative offset, got %d", chapter, i, streamByte.Offset)
			}
		}

		// Validate that sum of all runtime values equals audio duration
		totalRuntime := 0.0
		for _, streamByte := range fileData.Bytes {
			totalRuntime += streamByte.Runtime
		}
		t.Logf("Chapter %d: Total runtime from stream bytes: %.2fs", chapter, totalRuntime)

		totalFiles++
		totalTimestamps += len(timestamps)
		t.Logf("Successfully processed HLS file for chapter %d: %s", chapter, fileData.File.FileName)
		t.Logf("Generated %d stream bytes entries for chapter %d", len(fileData.Bytes), chapter)
	}

	t.Logf("Successfully processed %d HLS files", totalFiles)
	t.Logf("Total timestamps processed: %d", totalTimestamps)
}

func TestHLSDatabaseInsert(t *testing.T) {
	// Set up test environment - SQLite only
	os.Setenv("FCBH_DATASET_DB", "/Users/jrstear/git/artie/bible_brain/timestamp/update/test_data/init.db")

	ctx := context.Background()

	// Create test database connection to SQLite
	conn := db.NewDBAdapter(ctx, "test_data/init.db")
	defer conn.Close()

	// Clean up any previous test entries
	cleanupTestHLSData(conn, "TESTHLS002")

	// Create test HLS data
	hlsData := HLSData{
		Fileset: HLSFileset{
			ID:             "TESTHLS002",
			SetTypeCode:    "audio_stream",
			SetSizeCode:    "NT",
			ModeID:         3, // Test with mode_id = 3 (audio)
			HashID:         "testhash456",
			BibleID:        "ENGNIV",
			LicenseGroupID: nil,   // Test with no license group
			PublishedSNM:   false, // Test with not published
			CreatedAt:      "2024-01-01 12:00:00",
			UpdatedAt:      "2024-01-01 12:00:00",
		},
		FileGroups: []HLSFileGroup{
			{
				File: HLSFile{
					HashID:     "testhash456",
					BookID:     "JHN",
					ChapterNum: 1,
					FileName:   "JHN_001.m3u8",
					FileSize:   1024,
					CreatedAt:  "2024-01-01 12:00:00",
					UpdatedAt:  "2024-01-01 12:00:00",
				},
				Bandwidths: []HLSStreamBandwidth{
					{
						BibleFileID: 1, // Will be set by the insert process
						FileName:    "JHN_001-64kbs.m3u8",
						Bandwidth:   64000,
						Codec:       "mp4a.40.2",
						Stream:      1,
						CreatedAt:   "2024-01-01 12:00:00",
						UpdatedAt:   "2024-01-01 12:00:00",
					},
				},
				Bytes: []HLSStreamBytes{
					{
						StreamBandwidthID: 1, // Will be set by the insert process
						Runtime:           2.5,
						Bytes:             2000,
						Offset:            0,
						TimestampID:       1,
						CreatedAt:         "2024-01-01 12:00:00",
						UpdatedAt:         "2024-01-01 12:00:00",
					},
					{
						StreamBandwidthID: 1,
						Runtime:           2.5,
						Bytes:             2000,
						Offset:            2000,
						TimestampID:       2,
						CreatedAt:         "2024-01-01 12:00:00",
						UpdatedAt:         "2024-01-01 12:00:00",
					},
				},
			},
		},
	}

	// Insert HLS data using SQLite function
	status := insertHLSDataIntoSQLite(conn, hlsData)
	if status != nil {
		t.Fatalf("Failed to insert HLS data: %v", status)
	}

	// Verify the data was inserted correctly
	var count int
	query := "SELECT COUNT(*) FROM bible_filesets WHERE id = 'TESTHLS002'"
	err := conn.DB.QueryRow(query).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query HLS fileset: %v", err)
	}
	if count == 0 {
		t.Error("HLS fileset was not created")
	}

	// Verify HLS files
	query = "SELECT COUNT(*) FROM bible_files WHERE hash_id = 'testhash456'"
	err = conn.DB.QueryRow(query).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query HLS files: %v", err)
	}
	if count == 0 {
		t.Error("No HLS files were created")
	}

	// Verify stream bandwidths
	query = "SELECT COUNT(*) FROM bible_file_stream_bandwidths WHERE bible_file_id IN (SELECT id FROM bible_files WHERE hash_id = 'testhash456')"
	err = conn.DB.QueryRow(query).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query stream bandwidths: %v", err)
	}
	if count == 0 {
		t.Error("No stream bandwidths were created")
	}

	// Verify stream bytes
	query = "SELECT COUNT(*) FROM bible_file_stream_bytes WHERE stream_bandwidth_id IN (SELECT id FROM bible_file_stream_bandwidths WHERE bible_file_id IN (SELECT id FROM bible_files WHERE hash_id = 'testhash456'))"
	err = conn.DB.QueryRow(query).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query stream bytes: %v", err)
	}
	if count == 0 {
		t.Error("No stream bytes were created")
	}

	t.Logf("Successfully inserted HLS data for fileset: %s", hlsData.Fileset.ID)
	t.Logf("Created %d files, %d bandwidths, %d bytes", 1, 1, 2)
}

func TestHLSIntegration(t *testing.T) {
	// Set up test environment
	os.Setenv("FCBH_DATASET_DB", "/Users/jrstear/git/artie/bible_brain/timestamp/update/test_data/init.db")
	os.Setenv("FCBH_DATASET_FILES", "/Users/jrstear/git/artie/bible_brain/timestamp/update/test_data")
	os.Setenv("DBP_MYSQL_DSN", "root:@tcp(localhost)/jrs")

	ctx := context.Background()

	// Create test request
	req := request.Request{
		BibleId: "ENGNIV",
		UpdateDBP: request.UpdateDBP{
			Timestamps: "ENGNIVN1DA",
			HLS:        "TESTHLS002",
		},
	}

	// Create database connection
	conn := db.NewDBAdapter(ctx, "test_data/init.db")
	defer conn.Close()

	// Create UpdateTimestamps instance
	updateTimestamps := NewUpdateTimestamps(ctx, req, conn)

	// Process HLS
	status := updateTimestamps.ProcessHLS("TESTHLS002", "ENGNIV")
	if status != nil {
		t.Fatalf("Failed to process HLS: %v", status)
	}

	t.Logf("Successfully processed HLS integration test")
}

func TestHLSStreamGenerationForDAFilesetInSQLite(t *testing.T) {
	// Set up test environment - use SQLite only
	os.Setenv("FCBH_DATASET_DB", "/Users/jrstear/git/artie/bible_brain/timestamp/update/test_data/init.db")
	os.Setenv("FCBH_DATASET_FILES", "/Users/jrstear/tmp/artie/files")

	ctx := context.Background()

	// Create test database connection to SQLite
	conn := db.NewDBAdapter(ctx, "test_data/init.db")
	defer conn.Close()

	// Clean up any previous test entries
	cleanupTestHLSData(conn, "ENGNIVN1SA")

	// Create HLS processor
	processor := NewLocalHLSProcessor(ctx, "ENGNIV", "ENGNIVN1DA")

	// Process John chapters 1 and 2
	chapters := []int{1, 2}
	totalFiles := 0
	totalTimestamps := 0

	// Generate HLS data for each chapter
	var allHLSData HLSData
	now := time.Now().Format("2006-01-02 15:04:05")

	// Create the HLS fileset
	allHLSData.Fileset = HLSFileset{
		ID:             "ENGNIVN1SA",
		SetTypeCode:    "audio_stream",
		SetSizeCode:    "NT",
		HashID:         generateHashID("ENGNIVN1SA", "audio_stream"),
		BibleID:        "ENGNIV",
		LicenseGroupID: nil,   // Test with no license group
		PublishedSNM:   false, // Test with not published
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	for _, chapter := range chapters {
		// Get timestamps for this chapter from bible_file_timestamps table
		query := `SELECT bft.id, bft.verse_start, bft.timestamp, bft.timestamp_end, bf.file_name 
				  FROM bible_file_timestamps bft 
				  JOIN bible_files bf ON bft.bible_file_id = bf.id 
				  WHERE bf.book_id = ? AND bf.chapter_start = ? 
				  ORDER BY bft.verse_sequence`

		rows, err := conn.DB.Query(query, "JHN", chapter)
		if err != nil {
			t.Fatalf("Failed to query timestamps for JHN chapter %d: %v", chapter, err)
		}

		var timestamps []Timestamp
		verseSeq := 1
		for rows.Next() {
			var id int64
			var verseStart string
			var beginTS, endTS float64
			var fileName string

			err := rows.Scan(&id, &verseStart, &beginTS, &endTS, &fileName)
			if err != nil {
				t.Fatalf("Failed to scan timestamp for JHN chapter %d: %v", chapter, err)
			}

			timestamps = append(timestamps, Timestamp{
				TimestampId: id,
				VerseStr:    verseStart,
				BeginTS:     beginTS,
				EndTS:       endTS,
				AudioFile:   fileName,
				VerseSeq:    verseSeq,
			})
			verseSeq++
		}
		rows.Close()

		if len(timestamps) == 0 {
			t.Logf("No timestamps found for JHN chapter %d, skipping", chapter)
			continue
		}

		// Get the audio file name for this chapter
		audioFile := fmt.Sprintf("B04___%02d_John________ENGNIVN1DA.mp3", chapter)

		// Process the file
		fileData, err := processor.ProcessFile(audioFile, timestamps)
		if err != nil {
			t.Fatalf("Failed to process HLS file for chapter %d: %v", chapter, err)
		}

		// Set file metadata
		fileData.File.HashID = allHLSData.Fileset.HashID
		fileData.File.BookID = "JHN"
		fileData.File.ChapterNum = chapter
		fileData.File.CreatedAt = now
		fileData.File.UpdatedAt = now
		fileData.File.FileName = fmt.Sprintf("B04___%02d_John________ENGNIVN1SA.m3u8", chapter)

		// Update bandwidth metadata
		for i := range fileData.Bandwidths {
			fileData.Bandwidths[i].CreatedAt = now
			fileData.Bandwidths[i].UpdatedAt = now
			fileData.Bandwidths[i].FileName = fmt.Sprintf("B04___%02d_John________ENGNIVN1SA-64kbs.m3u8", chapter)
		}

		// Update bytes metadata
		for i := range fileData.Bytes {
			fileData.Bytes[i].CreatedAt = now
			fileData.Bytes[i].UpdatedAt = now
		}

		// Create file group for this chapter
		fileGroup := HLSFileGroup{
			File: HLSFile{
				HashID:     allHLSData.Fileset.HashID,
				BookID:     "JHN",
				ChapterNum: chapter,
				FileName:   fileData.File.FileName,
				FileSize:   fileData.File.FileSize,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			Bandwidths: fileData.Bandwidths,
			Bytes:      fileData.Bytes,
		}

		// Add file group to HLS data
		allHLSData.FileGroups = append(allHLSData.FileGroups, fileGroup)

		totalFiles++
		totalTimestamps += len(timestamps)
		t.Logf("Successfully processed HLS file for chapter %d: %s", chapter, fileData.File.FileName)
		t.Logf("Generated %d stream bytes entries for chapter %d", len(fileData.Bytes), chapter)
	}

	// Insert HLS data into SQLite database
	status := insertHLSDataIntoSQLite(conn, allHLSData)
	if status != nil {
		t.Fatalf("Failed to insert HLS data into SQLite: %v", status)
	}

	// Verify the data was inserted
	var count int
	query := "SELECT COUNT(*) FROM bible_filesets WHERE id = 'ENGNIVN1SA'"
	err := conn.DB.QueryRow(query).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query HLS fileset: %v", err)
	}

	if count == 0 {
		t.Error("HLS fileset was not created")
	}

	// Verify HLS files
	query = "SELECT COUNT(*) FROM bible_files WHERE hash_id = ?"
	err = conn.DB.QueryRow(query, allHLSData.Fileset.HashID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query HLS files: %v", err)
	}

	if count == 0 {
		t.Error("No HLS files were created")
	}

	// Verify stream bandwidths - check for any bandwidths with the HLS fileset files
	query = "SELECT COUNT(*) FROM bible_file_stream_bandwidths WHERE bible_file_id IN (SELECT id FROM bible_files WHERE hash_id = ?)"
	err = conn.DB.QueryRow(query, allHLSData.Fileset.HashID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query stream bandwidths: %v", err)
	}

	if count == 0 {
		t.Error("No stream bandwidths were created")
	}

	// Verify stream bytes - check for any bytes with the HLS fileset files
	query = "SELECT COUNT(*) FROM bible_file_stream_bytes WHERE stream_bandwidth_id IN (SELECT id FROM bible_file_stream_bandwidths WHERE bible_file_id IN (SELECT id FROM bible_files WHERE hash_id = ?))"
	err = conn.DB.QueryRow(query, allHLSData.Fileset.HashID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query stream bytes: %v", err)
	}

	if count == 0 {
		t.Error("No stream bytes were created")
	}

	t.Logf("Successfully generated HLS streams for ENGNIVN1DA fileset in SQLite")
	t.Logf("Created HLS fileset: ENGNIVN1SA with hash_id: %s", allHLSData.Fileset.HashID)
	t.Logf("Created %d HLS files", totalFiles)
	t.Logf("Total timestamps processed: %d", totalTimestamps)
}

// insertHLSDataIntoSQLite inserts HLS data into SQLite database
func insertHLSDataIntoSQLite(conn db.DBAdapter, hlsData HLSData) *log.Status {
	ctx := context.Background()

	// Insert fileset
	filesetQuery := `INSERT INTO bible_filesets (id, hash_id, asset_id, set_type_code, set_size_code, mode_id, license_group_id, published_snm, hidden, created_at, updated_at) 
					  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := conn.DB.Exec(filesetQuery,
		hlsData.Fileset.ID,
		hlsData.Fileset.HashID,
		"dbp-prod",
		hlsData.Fileset.SetTypeCode,
		hlsData.Fileset.SetSizeCode,
		hlsData.Fileset.ModeID,
		hlsData.Fileset.LicenseGroupID,
		hlsData.Fileset.PublishedSNM,
		0,
		hlsData.Fileset.CreatedAt,
		hlsData.Fileset.UpdatedAt)
	if err != nil {
		return log.Error(ctx, 500, err, "Failed to insert HLS fileset")
	}

	// Process each file group
	for _, fileGroup := range hlsData.FileGroups {
		// Insert file
		fileQuery := `INSERT INTO bible_files (hash_id, book_id, chapter_start, file_name, file_size, created_at, updated_at) 
					  VALUES (?, ?, ?, ?, ?, ?, ?)`

		result, err := conn.DB.Exec(fileQuery,
			fileGroup.File.HashID,
			fileGroup.File.BookID,
			fileGroup.File.ChapterNum,
			fileGroup.File.FileName,
			fileGroup.File.FileSize,
			fileGroup.File.CreatedAt,
			fileGroup.File.UpdatedAt)
		if err != nil {
			return log.Error(ctx, 500, err, "Failed to insert HLS file")
		}

		// Get the file ID
		fileID, err := result.LastInsertId()
		if err != nil {
			return log.Error(ctx, 500, err, "Failed to get file ID")
		}

		// Insert bandwidths for this file
		bandwidthIDMap := make(map[string]int64)
		for _, bandwidth := range fileGroup.Bandwidths {
			bandwidthQuery := `INSERT INTO bible_file_stream_bandwidths (bible_file_id, file_name, bandwidth, codec, stream, created_at, updated_at) 
							   VALUES (?, ?, ?, ?, ?, ?, ?)`

			result, err := conn.DB.Exec(bandwidthQuery,
				fileID,
				bandwidth.FileName,
				bandwidth.Bandwidth,
				bandwidth.Codec,
				bandwidth.Stream,
				bandwidth.CreatedAt,
				bandwidth.UpdatedAt)
			if err != nil {
				return log.Error(ctx, 500, err, "Failed to insert stream bandwidth")
			}

			// Get bandwidth ID
			bandwidthID, err := result.LastInsertId()
			if err != nil {
				return log.Error(ctx, 500, err, "Failed to get bandwidth ID")
			}
			bandwidthIDMap[bandwidth.FileName] = bandwidthID
		}

		// Insert bytes for this file
		for _, streamByte := range fileGroup.Bytes {
			// Use the first bandwidth ID for this file (single bandwidth case)
			var bandwidthID int64
			for _, id := range bandwidthIDMap {
				bandwidthID = id
				break
			}

			bytesQuery := `INSERT INTO bible_file_stream_bytes (stream_bandwidth_id, runtime, bytes, offset, timestamp_id, created_at, updated_at) 
						   VALUES (?, ?, ?, ?, ?, ?, ?)`

			_, err := conn.DB.Exec(bytesQuery,
				bandwidthID,
				streamByte.Runtime,
				streamByte.Bytes,
				streamByte.Offset,
				streamByte.TimestampID,
				streamByte.CreatedAt,
				streamByte.UpdatedAt)
			if err != nil {
				return log.Error(ctx, 500, err, "Failed to insert stream bytes")
			}
		}
	}

	return nil
}

// cleanupTestHLSData removes test HLS data from previous runs
func cleanupTestHLSData(conn db.DBAdapter, filesetID string) {
	// Delete in reverse order due to foreign key constraints

	// Delete stream bytes
	_, err := conn.DB.Exec(`DELETE FROM bible_file_stream_bytes 
							WHERE stream_bandwidth_id IN (
								SELECT id FROM bible_file_stream_bandwidths 
								WHERE bible_file_id IN (
									SELECT id FROM bible_files 
									WHERE hash_id IN (
										SELECT hash_id FROM bible_filesets WHERE id = ?
									)
								)
							)`, filesetID)
	if err != nil {
		fmt.Printf("Warning: Failed to cleanup stream bytes for %s: %v\n", filesetID, err)
	}

	// Delete stream bandwidths
	_, err = conn.DB.Exec(`DELETE FROM bible_file_stream_bandwidths 
						   WHERE bible_file_id IN (
							   SELECT id FROM bible_files 
							   WHERE hash_id IN (
								   SELECT hash_id FROM bible_filesets WHERE id = ?
							   )
						   )`, filesetID)
	if err != nil {
		fmt.Printf("Warning: Failed to cleanup stream bandwidths for %s: %v\n", filesetID, err)
	}

	// Delete files
	_, err = conn.DB.Exec(`DELETE FROM bible_files 
						   WHERE hash_id IN (
							   SELECT hash_id FROM bible_filesets WHERE id = ?
						   )`, filesetID)
	if err != nil {
		fmt.Printf("Warning: Failed to cleanup files for %s: %v\n", filesetID, err)
	}

	// Delete fileset
	_, err = conn.DB.Exec(`DELETE FROM bible_filesets WHERE id = ?`, filesetID)
	if err != nil {
		fmt.Printf("Warning: Failed to cleanup fileset %s: %v\n", filesetID, err)
	}
}
