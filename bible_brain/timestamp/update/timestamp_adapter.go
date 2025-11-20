package update

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	_ "github.com/go-sql-driver/mysql"
)

type DBPAdapter struct {
	ctx  context.Context
	conn *sql.DB
}

func NewDBPAdapter(ctx context.Context) (DBPAdapter, *log.Status) {
	var dbp DBPAdapter
	dbp.ctx = ctx
	var err error
	mysqlDSN := os.Getenv("DBP_MYSQL_DSN")
	mysqlDSN += "?clientFoundRows=true" // makes affected row count number matched, not number updated
	dbp.conn, err = sql.Open("mysql", mysqlDSN)
	if err != nil {
		return dbp, log.Error(dbp.ctx, 500, err, "Error connecting to dbp database")
	}
	err = dbp.conn.Ping()
	if err != nil {
		return dbp, log.Error(dbp.ctx, 500, err, "Connection to dbp database ping failed")
	}
	return dbp, nil
}

func (d *DBPAdapter) Close() {
	_ = d.conn.Close()
}

func (d *DBPAdapter) SelectHashId(filesetId string) (string, *log.Status) {
	var result string
	query := `SELECT hash_id FROM bible_filesets WHERE id = ?`
	rows, err := d.conn.Query(query, filesetId)
	if err != nil {
		return result, log.Error(d.ctx, 500, err, query)
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&result)
		if err != nil {
			return result, log.Error(d.ctx, 500, err, query)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.ctx, err, query)
	}
	return result, nil
}

// SelectAssetId gets the asset_id for a given fileset ID
func (d *DBPAdapter) SelectAssetId(filesetId string) (string, *log.Status) {
	var result string
	query := `SELECT asset_id FROM bible_filesets WHERE id = ?`
	rows, err := d.conn.Query(query, filesetId)
	if err != nil {
		return result, log.Error(d.ctx, 500, err, query)
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&result)
		if err != nil {
			return result, log.Error(d.ctx, 500, err, query)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.ctx, err, query)
	}
	return result, nil
}

func (d *DBPAdapter) SelectFileId(hashId string, bookId string, chapterNum int) (int64, string, *log.Status) {
	var result int64
	var filename string
	query := `SELECT distinct id, file_name FROM bible_files WHERE hash_id = ? AND book_id = ? and chapter_start = ?`
	rows, err := d.conn.Query(query, hashId, bookId, chapterNum)
	if err != nil {
		return result, filename, log.Error(d.ctx, 500, err, query)
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&result, &filename)
		if err != nil {
			return result, filename, log.Error(d.ctx, 500, err, query)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.ctx, err, query)
	}
	return result, filename, nil
}

func (d *DBPAdapter) SelectTimestamps(fileId int64) ([]Timestamp, *log.Status) {
	var result []Timestamp
	query := `SELECT id, verse_start, verse_end, verse_sequence, timestamp, timestamp_end 
		FROM bible_file_timestamps WHERE bible_file_id = ? 
		ORDER BY verse_sequence, id`
	rows, err := d.conn.Query(query, fileId)
	if err != nil {
		return result, log.Error(d.ctx, 500, err, query)
	}
	defer rows.Close()
	for rows.Next() {
		var tmpEndTS sql.NullFloat64
		var rec Timestamp
		err = rows.Scan(&rec.TimestampId, &rec.VerseStr, &rec.VerseEnd,
			&rec.VerseSeq, &rec.BeginTS, &tmpEndTS)
		if err != nil {
			return result, log.Error(d.ctx, 500, err, query)
		}
		// How should I handle verse end
		if tmpEndTS.Valid {
			rec.EndTS = tmpEndTS.Float64
		}
		result = append(result, rec)
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.ctx, err, query)
	}
	return result, nil
}

// GetFilesetDurations returns a map[book][chapter]durationSeconds derived from bible_files.duration
func (d *DBPAdapter) GetFilesetDurations(filesetID string) (map[string]map[int]float64, *log.Status) {
	results := make(map[string]map[int]float64)

	hashID, status := d.SelectHashId(filesetID)
	if status != nil {
		return nil, status
	}

	query := `
		SELECT bf.book_id, bf.chapter_start, bf.duration
		FROM bible_files bf
		WHERE bf.hash_id = ? AND bf.duration IS NOT NULL
	`

	rows, err := d.conn.Query(query, hashID)
	if err != nil {
		return nil, log.Error(d.ctx, 500, err, "Failed to query fileset durations")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			bookID   sql.NullString
			chapter  sql.NullInt64
			duration sql.NullInt64
		)

		if err := rows.Scan(&bookID, &chapter, &duration); err != nil {
			return nil, log.Error(d.ctx, 500, err, "Failed to scan fileset duration row")
		}

		if !bookID.Valid || !chapter.Valid || !duration.Valid {
			continue
		}

		bookMap, ok := results[bookID.String]
		if !ok {
			bookMap = make(map[int]float64)
			results[bookID.String] = bookMap
		}
		// Convert int to float64 for consistency with existing code
		bookMap[int(chapter.Int64)] = float64(duration.Int64)
	}

	if err := rows.Err(); err != nil {
		return nil, log.Error(d.ctx, 500, err, "Error iterating fileset durations")
	}

	return results, nil
}

// GetBibleFileDuration gets the duration field from bible_files for a specific file
func (d *DBPAdapter) GetBibleFileDuration(filesetID string, bookID string, chapterNum int, filename string) (*int, *log.Status) {
	hashID, status := d.SelectHashId(filesetID)
	if status != nil {
		return nil, status
	}

	query := `
		SELECT bf.duration
		FROM bible_files bf
		WHERE bf.hash_id = ? AND bf.book_id = ? AND bf.chapter_start = ? AND bf.file_name = ?
	`

	var duration sql.NullInt64
	err := d.conn.QueryRow(query, hashID, bookID, chapterNum, filename).Scan(&duration)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// File not found in database, return nil (no duration available)
			return nil, nil
		}
		return nil, log.Error(d.ctx, 500, err, "Failed to get bible_files.duration")
	}

	if !duration.Valid {
		// Duration is NULL in database
		return nil, nil
	}

	result := int(duration.Int64)
	return &result, nil
}

// GetFilesetTimestamps returns timestamps grouped by book/chapter along with an ordered chapter list
func (d *DBPAdapter) GetFilesetTimestamps(filesetID string) (map[string]map[int][]Timestamp, []db.Script, *log.Status) {
	result := make(map[string]map[int][]Timestamp)
	var chapters []db.Script

	hashID, status := d.SelectHashId(filesetID)
	if status != nil {
		return nil, nil, status
	}

	query := `
		SELECT DISTINCT book_id, chapter_start
		FROM bible_files
		WHERE hash_id = ?
		  AND book_id <> ''
		  AND chapter_start IS NOT NULL
		ORDER BY book_id, chapter_start
	`

	rows, err := d.conn.Query(query, hashID)
	if err != nil {
		return nil, nil, log.Error(d.ctx, 500, err, "Failed to query fileset chapters")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			bookID  sql.NullString
			chapter sql.NullInt64
		)

		if err := rows.Scan(&bookID, &chapter); err != nil {
			return nil, nil, log.Error(d.ctx, 500, err, "Failed to scan fileset chapter row")
		}

		if !bookID.Valid || !chapter.Valid {
			continue
		}

		chapterNum := int(chapter.Int64)
		fileID, filename, status := d.SelectFileId(hashID, bookID.String, chapterNum)
		if status != nil {
			return nil, nil, status
		}
		if fileID <= 0 {
			continue
		}

		timestamps, status := d.SelectTimestamps(fileID)
		if status != nil {
			return nil, nil, status
		}
		if len(timestamps) == 0 {
			continue
		}

		for i := range timestamps {
			timestamps[i].AudioFile = filename
		}

		bookMap, ok := result[bookID.String]
		if !ok {
			bookMap = make(map[int][]Timestamp)
			result[bookID.String] = bookMap
		}
		bookMap[chapterNum] = timestamps
		chapters = append(chapters, db.Script{BookId: bookID.String, ChapterNum: chapterNum})
	}

	if err := rows.Err(); err != nil {
		return nil, nil, log.Error(d.ctx, 500, err, "Error iterating fileset chapters")
	}

	sort.Slice(chapters, func(i, j int) bool {
		if chapters[i].BookId == chapters[j].BookId {
			return chapters[i].ChapterNum < chapters[j].ChapterNum
		}
		return chapters[i].BookId < chapters[j].BookId
	})

	return result, chapters, nil
}

func (d *DBPAdapter) InsertTimestamps(bibleFileId int64, timestamps []Timestamp) ([]Timestamp, int, *log.Status) {
	var rowCount int
	var mustInsert int
	for _, rec := range timestamps {
		if rec.TimestampId == 0 {
			mustInsert++
		}
	}
	if mustInsert > 0 {
		query := `INSERT INTO bible_file_timestamps (bible_file_id, verse_start, verse_end,
		timestamp, timestamp_end, verse_sequence) VALUES (?,?,?,?,?,?)`
		tx, err := d.conn.Begin()
		if err != nil {
			return timestamps, rowCount, log.Error(d.ctx, 500, err, query)
		}
		stmt, err := tx.Prepare(query)
		if err != nil {
			return timestamps, rowCount, log.Error(d.ctx, 500, err, query)
		}
		defer stmt.Close()
		var result sql.Result
		var count int64
		for i, rec := range timestamps {
			if rec.TimestampId == 0 {
				result, err = stmt.Exec(bibleFileId, rec.VerseStr, rec.VerseEnd, rec.BeginTS, rec.EndTS, rec.VerseSeq)
				if err != nil {
					return timestamps, rowCount, log.Error(d.ctx, 500, err, `Error while inserting dbp timestamp.`)
				}
				timestamps[i].TimestampId, err = result.LastInsertId()
				if err != nil {
					return timestamps, rowCount, log.Error(d.ctx, 500, err, `Error getting lastInsertId, while inserting Timestamps.`)
				}
				count, err = result.RowsAffected()
				if err != nil {
					return timestamps, rowCount, log.Error(d.ctx, 500, err, query)
				}
				rowCount += int(count)
			}
		}
		err = tx.Commit()
		if err != nil {
			return timestamps, rowCount, log.Error(d.ctx, 500, err, query)
		}
		if rowCount != mustInsert {
			return timestamps, rowCount, log.ErrorNoErr(d.ctx, 500,
				"Row count expected:", mustInsert, "Actual Count:", rowCount, query)
		}
	}
	return timestamps, rowCount, nil
}

// FindSAFilesetForBooks checks if any SA fileset has bible_files that reference DA timestamps for specific books
// Returns the SA fileset ID and list of affected book IDs
func (d *DBPAdapter) FindSAFilesetForBooks(daFilesetID string, bookIDs []string) (string, []string, *log.Status) {
	if len(bookIDs) == 0 {
		return "", nil, nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(bookIDs))
	args := make([]interface{}, len(bookIDs)+1)
	args[0] = daFilesetID
	for i, bookID := range bookIDs {
		placeholders[i] = "?"
		args[i+1] = bookID
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT bf_sa.id, bf_sa_file.book_id
		FROM bible_filesets bf_da
		JOIN bible_files bf_da_file ON bf_da.hash_id = bf_da_file.hash_id
		JOIN bible_file_timestamps ts ON bf_da_file.id = ts.bible_file_id
		JOIN bible_file_stream_bytes bytes ON ts.id = bytes.timestamp_id
		JOIN bible_file_stream_bandwidths bw ON bytes.stream_bandwidth_id = bw.id
		JOIN bible_files bf_sa_file ON bw.bible_file_id = bf_sa_file.id
		JOIN bible_filesets bf_sa ON bf_sa_file.hash_id = bf_sa.hash_id
		WHERE bf_da.id = ? AND bf_da_file.book_id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return "", nil, log.Error(d.ctx, 500, err, query)
	}
	defer rows.Close()

	var saFilesetID string
	affectedBooks := make(map[string]bool)
	for rows.Next() {
		var fsID, bookID string
		if err := rows.Scan(&fsID, &bookID); err != nil {
			return "", nil, log.Error(d.ctx, 500, err, "Error scanning SA fileset result")
		}
		saFilesetID = fsID
		affectedBooks[bookID] = true
	}

	if err = rows.Err(); err != nil {
		return "", nil, log.Error(d.ctx, 500, err, "Error iterating SA fileset rows")
	}

	// Convert map to slice
	bookList := make([]string, 0, len(affectedBooks))
	for book := range affectedBooks {
		bookList = append(bookList, book)
	}

	return saFilesetID, bookList, nil
}

// removeStreamBooksTx deletes bible_files for specific books/chapters in SA fileset
// CASCADE will automatically remove bandwidths and bytes
func (d *DBPAdapter) removeStreamBooksTx(tx *sql.Tx, saFilesetID string, chapters []db.Script) error {
	if saFilesetID == "" || len(chapters) == 0 {
		return nil
	}

	// Get hash_id for SA fileset
	var hashID string
	query := `SELECT hash_id FROM bible_filesets WHERE id = ?`
	err := tx.QueryRow(query, saFilesetID).Scan(&hashID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // Fileset doesn't exist, nothing to remove
		}
		return err
	}

	// Delete bible_files for each chapter
	deleteQuery := `DELETE FROM bible_files WHERE hash_id = ? AND book_id = ? AND chapter_start = ?`
	stmt, err := tx.Prepare(deleteQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, ch := range chapters {
		_, err = stmt.Exec(hashID, ch.BookId, ch.ChapterNum)
		if err != nil {
			return err
		}
	}

	return nil
}

// removeTimestampsForBooksTx deletes timestamps for specific books/chapters
func (d *DBPAdapter) removeTimestampsForBooksTx(tx *sql.Tx, filesetID string, chapters []db.Script) error {
	if len(chapters) == 0 {
		return nil
	}

	// Get hash_id for fileset
	var hashID string
	query := `SELECT hash_id FROM bible_filesets WHERE id = ?`
	err := tx.QueryRow(query, filesetID).Scan(&hashID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // Fileset doesn't exist, nothing to remove
		}
		return err
	}

	// Delete timestamps for each chapter
	deleteQuery := `DELETE FROM bible_file_timestamps 
		WHERE bible_file_id IN (
			SELECT f.id FROM bible_files f
			WHERE f.hash_id = ? AND f.book_id = ? AND f.chapter_start = ?
		)`
	stmt, err := tx.Prepare(deleteQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, ch := range chapters {
		_, err = stmt.Exec(hashID, ch.BookId, ch.ChapterNum)
		if err != nil {
			return err
		}
	}

	return nil
}

// insertTimestampsTx inserts timestamps within a transaction
func (d *DBPAdapter) insertTimestampsTx(tx *sql.Tx, bibleFileId int64, timestamps []Timestamp) ([]Timestamp, error) {
	if len(timestamps) == 0 {
		return timestamps, nil
	}

	query := `INSERT INTO bible_file_timestamps (bible_file_id, verse_start, verse_end,
		timestamp, timestamp_end, verse_sequence) VALUES (?,?,?,?,?,?)`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return timestamps, err
	}
	defer stmt.Close()

	for i, rec := range timestamps {
		if rec.TimestampId == 0 {
			result, err := stmt.Exec(bibleFileId, rec.VerseStr, rec.VerseEnd, rec.BeginTS, rec.EndTS, rec.VerseSeq)
			if err != nil {
				return timestamps, err
			}
			timestamps[i].TimestampId, err = result.LastInsertId()
			if err != nil {
				return timestamps, err
			}
		}
	}

	return timestamps, nil
}

// updateFilesetTimingEstTagTx inserts or updates the timing_est_err tag within a transaction
func (d *DBPAdapter) updateFilesetTimingEstTagTx(tx *sql.Tx, hashId, timingEstErr string) error {
	query := `SELECT description FROM bible_fileset_tags WHERE hash_id = ? AND name = 'timing_est_err'`
	var currEstErr string
	err := tx.QueryRow(query, hashId).Scan(&currEstErr)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if errors.Is(err, sql.ErrNoRows) {
		query = `INSERT INTO bible_fileset_tags (hash_id, name, description, admin_only, iso, language_id)
		VALUES (?, 'timing_est_err', ?, 0, 'eng', 6414)`
		_, err = tx.Exec(query, hashId, timingEstErr)
		return err
	} else if currEstErr != timingEstErr {
		query = `UPDATE bible_fileset_tags SET description = ? WHERE hash_id = ? AND name = 'timing_est_err'`
		_, err = tx.Exec(query, timingEstErr, hashId)
		return err
	}

	return nil
}

// DurationUpdate represents a single duration update to be applied
type DurationUpdate struct {
	FilesetID string
	BookID    string
	Chapter   int
	Filename  string
	Duration  int
}

// getAudioDuration uses FFmpeg to get the actual audio duration
func (d *DBPAdapter) getAudioDuration(audioPath string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", audioPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %v", err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %v", err)
	}

	return duration, nil
}

// validateDuration validates duration and returns authoritative duration, current db duration, and error
func (d *DBPAdapter) validateDuration(audioPath string, timestamps []Timestamp, filesetID, bookID string, chapterNum int, filename string) (int, *int, *log.Status) {
	// Get current duration from database
	dbDuration, status := d.GetBibleFileDuration(filesetID, bookID, chapterNum, filename)
	if status != nil {
		return 0, nil, status
	}

	// Get audio duration from format metadata
	audioDuration, err := d.getAudioDuration(audioPath)
	if err != nil {
		return 0, nil, log.Error(d.ctx, 500, err, fmt.Sprintf("Failed to get audio duration for %s", audioPath))
	}

	// Calculate sum of verse durations
	totalVerseDuration := 0.0
	for _, timestamp := range timestamps {
		totalVerseDuration += (timestamp.EndTS - timestamp.BeginTS)
	}

	// Round all values to int (seconds) for comparison
	roundedFormatDuration := int(math.Round(audioDuration))
	roundedSumVerses := int(math.Round(totalVerseDuration))

	// Critical validation: Format duration must be within 1 second of sum of verses
	// Allow 1 second tolerance to allow for possible rounding difference in upstream processes
	if math.Abs(float64(roundedFormatDuration-roundedSumVerses)) > 1 {
		return 0, nil, log.ErrorNoErr(d.ctx, 500, fmt.Sprintf("Audio duration mismatch for %s %s %d %s: format(%ds) vs sum_verses(%ds) - difference exceeds 1 second",
			filesetID, bookID, chapterNum, filename, roundedFormatDuration, roundedSumVerses))
	}

	// Format and sum agree - use format as authoritative
	authoritativeDuration := roundedFormatDuration

	// Note: If dbDuration differs from authoritativeDuration, we'll update it later
	// We don't fail here because format and sum agree, so we trust that value

	return authoritativeDuration, dbDuration, nil
}

// ProcessTimestamps processes timestamps for specific books/chapters in a single transaction
// It removes affected SA files, removes/inserts DA timestamps, updates the timing_est_err tag, and updates durations
func (d *DBPAdapter) ProcessTimestamps(daFilesetID, timingEstErr, bibleID string, chapters []db.Script, timestampsData map[string]map[int][]Timestamp) *log.Status {
	// Extract unique book IDs
	bookIDs := make(map[string]bool)
	for _, ch := range chapters {
		bookIDs[ch.BookId] = true
	}
	bookList := make([]string, 0, len(bookIDs))
	for bookID := range bookIDs {
		bookList = append(bookList, bookID)
	}

	// Check for SA fileset that references these books
	saFilesetID, affectedBooks, status := d.FindSAFilesetForBooks(daFilesetID, bookList)
	if status != nil {
		return status
	}

	// Get hash_id for DA fileset
	hashID, status := d.SelectHashId(daFilesetID)
	if status != nil {
		return status
	}

	// Validate durations and collect updates (before transaction)
	filesDir := os.Getenv("FCBH_DATASET_FILES")
	if filesDir == "" {
		filesDir = "/tmp/artie/files"
	}
	filesDir = filepath.Join(filesDir, bibleID, daFilesetID)

	var durationUpdates []DurationUpdate

	// Validate durations for each chapter
	for _, ch := range chapters {
		// Get timestamps for this chapter
		if bookTimestamps, ok := timestampsData[ch.BookId]; ok {
			if chapterTimestamps, ok := bookTimestamps[ch.ChapterNum]; ok && len(chapterTimestamps) > 0 {
				// Get filename from database
				_, filename, status := d.SelectFileId(hashID, ch.BookId, ch.ChapterNum)
				if status != nil {
					return status
				}
				if filename == "" {
					continue
				}

				// Construct audio file path
				audioPath := filepath.Join(filesDir, filename)

				// Validate duration (this will fail if mismatches found)
				authoritativeDuration, dbDuration, status := d.validateDuration(audioPath, chapterTimestamps, daFilesetID, ch.BookId, ch.ChapterNum, filename)
				if status != nil {
					return status
				}

				// Collect update if needed
				if dbDuration == nil || *dbDuration != authoritativeDuration {
					durationUpdates = append(durationUpdates, DurationUpdate{
						FilesetID: daFilesetID,
						BookID:    ch.BookId,
						Chapter:   ch.ChapterNum,
						Filename:  filename,
						Duration:  authoritativeDuration,
					})
				}
			}
		}
	}

	// Start transaction
	tx, err := d.conn.Begin()
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to begin timestamps transaction")
	}
	defer tx.Rollback()

	// Remove SA files for affected books (if any)
	if saFilesetID != "" && len(affectedBooks) > 0 {
		// Filter chapters to only those in affected books
		affectedChapters := make([]db.Script, 0)
		for _, ch := range chapters {
			for _, affectedBook := range affectedBooks {
				if ch.BookId == affectedBook {
					affectedChapters = append(affectedChapters, ch)
					break
				}
			}
		}
		if len(affectedChapters) > 0 {
			err = d.removeStreamBooksTx(tx, saFilesetID, affectedChapters)
			if err != nil {
				return log.Error(d.ctx, 500, err, "Failed to remove SA files for affected books")
			}
		}
	}

	// Remove existing DA timestamps for these books
	err = d.removeTimestampsForBooksTx(tx, daFilesetID, chapters)
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to remove DA timestamps for books")
	}

	// Insert new timestamps for each chapter
	for _, ch := range chapters {
		// Get file ID for this chapter
		fileID, _, status := d.SelectFileId(hashID, ch.BookId, ch.ChapterNum)
		if status != nil {
			return status
		}
		if fileID <= 0 {
			continue // Skip if no file found
		}

		// Get timestamps for this chapter from the map
		if bookTimestamps, ok := timestampsData[ch.BookId]; ok {
			if chapterTimestamps, ok := bookTimestamps[ch.ChapterNum]; ok && len(chapterTimestamps) > 0 {
				// Insert timestamps within transaction
				_, err = d.insertTimestampsTx(tx, fileID, chapterTimestamps)
				if err != nil {
					return log.Error(d.ctx, 500, err, "Failed to insert timestamps for chapter")
				}
			}
		}
	}

	// Update timing_est_err tag
	err = d.updateFilesetTimingEstTagTx(tx, hashID, timingEstErr)
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to update timing_est_err tag")
	}

	// Update durations within the same transaction
	if len(durationUpdates) > 0 {
		err = d.updateBibleFileDurationsTx(tx, durationUpdates, hashID)
		if err != nil {
			return log.Error(d.ctx, 500, err, "Failed to update durations")
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to commit timestamps transaction")
	}

	return nil
}

// updateBibleFileDurationsTx updates multiple duration fields within a transaction
func (d *DBPAdapter) updateBibleFileDurationsTx(tx *sql.Tx, updates []DurationUpdate, hashID string) error {
	query := `
		UPDATE bible_files 
		SET duration = ?, updated_at = CURRENT_TIMESTAMP
		WHERE hash_id = ? AND book_id = ? AND chapter_start = ? AND file_name = ?
	`

	for _, update := range updates {
		result, err := tx.Exec(query, update.Duration, hashID, update.BookID, update.Chapter, update.Filename)
		if err != nil {
			return fmt.Errorf("failed to update bible_files.duration for %s %s %d %s: %v",
				update.FilesetID, update.BookID, update.Chapter, update.Filename, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %v", err)
		}

		if rowsAffected == 0 {
			log.Warn(d.ctx, fmt.Sprintf("No bible_files record found to update: %s %s %d %s",
				update.FilesetID, update.BookID, update.Chapter, update.Filename))
		} else {
			log.Info(d.ctx, fmt.Sprintf("Updated bible_files.duration to %d for %s %s %d %s",
				update.Duration, update.FilesetID, update.BookID, update.Chapter, update.Filename))
		}
	}

	return nil
}
