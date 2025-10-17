package update

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
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

// ProcessTimestampsWithTag processes timestamps for specific books/chapters in a single transaction
// It removes affected SA files, removes/inserts DA timestamps, and updates the timing_est_err tag
func (d *DBPAdapter) ProcessTimestampsWithTag(daFilesetID, timingEstErr string, chapters []db.Script, timestampsData map[string]map[int][]Timestamp) *log.Status {
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

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to commit timestamps transaction")
	}

	return nil
}

// SelectBookChapterFromDBP gets chapters from MySQL database for a specific fileset
func (d *DBPAdapter) SelectBookChapterFromDBP(filesetId string) ([]db.Script, *log.Status) {
	var result []db.Script

	// Get the hash_id for the fileset
	hashId, status := d.SelectHashId(filesetId)
	if status != nil {
		return nil, status
	}

	query := `SELECT DISTINCT bf.book_id, bf.chapter_start, bf.chapter_end
		FROM bible_files bf 
		WHERE bf.hash_id = ?
		ORDER BY bf.book_id, bf.chapter_start`

	rows, err := d.conn.Query(query, hashId)
	if err != nil {
		return nil, log.Error(d.ctx, 500, err, query)
	}
	defer rows.Close()

	for rows.Next() {
		var script db.Script
		var chapterEnd sql.NullInt32
		err = rows.Scan(&script.BookId, &script.ChapterNum, &chapterEnd)
		if err != nil {
			return nil, log.Error(d.ctx, 500, err, query)
		}
		if chapterEnd.Valid {
			script.ChapterEnd = int(chapterEnd.Int32)
		}
		result = append(result, script)
	}

	return result, nil
}

// HLS Data Structures

type HLSFileset struct {
	ID             string
	SetTypeCode    string
	SetSizeCode    string
	ModeID         int
	HashID         string
	AssetID        string
	BibleID        string
	LicenseGroupID *int
	PublishedSNM   bool
	CreatedAt      string
	UpdatedAt      string
}

type HLSFile struct {
	ID         int64
	HashID     string
	BookID     string
	ChapterNum int
	FileName   string
	FileSize   int64
	Duration   int // Duration in seconds (rounded to int)
	CreatedAt  string
	UpdatedAt  string
}

type HLSStreamBandwidth struct {
	ID               int64
	BibleFileID      int64
	FileName         string
	Bandwidth        int
	ResolutionWidth  *int
	ResolutionHeight *int
	Codec            string
	Stream           int
	CreatedAt        string
	UpdatedAt        string
}

type HLSStreamBytes struct {
	ID                int64
	StreamBandwidthID int64
	Runtime           float64
	Bytes             int64
	Offset            int64
	TimestampID       int64
	CreatedAt         string
	UpdatedAt         string
}

type HLSData struct {
	Fileset    HLSFileset
	FileGroups []HLSFileGroup
}

type HLSFileGroup struct {
	File       HLSFile
	Bandwidths []HLSStreamBandwidth
	Bytes      []HLSStreamBytes
}

// HLS Database Operations

func (d *DBPAdapter) RemoveHLSFileset(filesetID string) *log.Status {
	// Get hash_id for the fileset
	var hashID string
	query := `SELECT hash_id FROM bible_filesets WHERE id = ?`
	err := d.conn.QueryRow(query, filesetID).Scan(&hashID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// Fileset doesn't exist, that's fine
			return nil
		}
		return log.Error(d.ctx, 500, err, "Failed to get hash_id for fileset: "+filesetID)
	}

	// Start transaction
	tx, err := d.conn.Begin()
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to begin transaction")
	}
	defer tx.Rollback()

	// Delete in reverse dependency order
	// 1. Delete stream bytes
	_, err = tx.Exec(`DELETE FROM bible_file_stream_bytes WHERE stream_bandwidth_id IN (SELECT id FROM bible_file_stream_bandwidths WHERE bible_file_id IN (SELECT id FROM bible_files WHERE hash_id = ?))`, hashID)
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to delete HLS stream bytes")
	}

	// 2. Delete stream bandwidths
	_, err = tx.Exec(`DELETE FROM bible_file_stream_bandwidths WHERE bible_file_id IN (SELECT id FROM bible_files WHERE hash_id = ?)`, hashID)
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to delete HLS stream bandwidths")
	}

	// 3. Delete bible files
	_, err = tx.Exec(`DELETE FROM bible_files WHERE hash_id = ?`, hashID)
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to delete HLS bible files")
	}

	// 4. Delete fileset connections
	_, err = tx.Exec(`DELETE FROM bible_fileset_connections WHERE hash_id = ?`, hashID)
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to delete HLS fileset connections")
	}

	// 5. Delete fileset tags
	_, err = tx.Exec(`DELETE FROM bible_fileset_tags WHERE hash_id = ?`, hashID)
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to delete HLS fileset tags")
	}

	// 6. Delete the fileset
	_, err = tx.Exec(`DELETE FROM bible_filesets WHERE hash_id = ?`, hashID)
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to delete HLS fileset")
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to commit HLS removal transaction")
	}

	return nil
}

// SelectFATimestampsFromDBP gets timestamps from MySQL database with actual timestamp IDs
func (d *DBPAdapter) SelectFATimestampsFromDBP(bookId string, chapter int, filesetId string) ([]Timestamp, *log.Status) {
	// Get the hash_id for the fileset
	hashId, status := d.SelectHashId(filesetId)
	if status != nil {
		return nil, status
	}

	// Get the file ID and filename for this book/chapter
	fileId, filename, status := d.SelectFileId(hashId, bookId, chapter)
	if status != nil {
		return nil, status
	}

	if fileId <= 0 {
		return []Timestamp{}, nil // No file found
	}

	// Get timestamps for this file
	timestamps, status := d.SelectTimestamps(fileId)
	if status != nil {
		return nil, status
	}

	// Set the AudioFile for each timestamp
	for i := range timestamps {
		timestamps[i].AudioFile = filename
	}

	return timestamps, nil
}

// SelectFilesetLicenseInfo gets mode_id, license_group_id and published_snm from a fileset
func (d *DBPAdapter) SelectFilesetLicenseInfo(filesetId string) (int, *int, bool, *log.Status) {
	query := `SELECT mode_id, license_group_id, published_snm FROM bible_filesets WHERE id = ?`

	var modeID int
	var licenseGroupID *int
	var publishedSNM bool

	err := d.conn.QueryRow(query, filesetId).Scan(&modeID, &licenseGroupID, &publishedSNM)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil, false, log.ErrorNoErr(d.ctx, 404, "Fileset not found: "+filesetId)
		}
		return 0, nil, false, log.Error(d.ctx, 500, err, "Failed to query fileset license info")
	}

	return modeID, licenseGroupID, publishedSNM, nil
}

func (d *DBPAdapter) InsertHLSData(hlsData HLSData) *log.Status {
	// First remove any existing HLS data for this fileset
	status := d.RemoveHLSFileset(hlsData.Fileset.ID)
	if status != nil {
		return status
	}

	// Start transaction
	tx, err := d.conn.Begin()
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to begin transaction")
	}
	defer tx.Rollback()

	// Insert fileset
	isSA := strings.HasSuffix(strings.ToUpper(hlsData.Fileset.ID), "SA")
	_, err = d.insertHLSFilesetTx(tx, hlsData.Fileset, isSA)
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to insert HLS fileset")
	}

	// Process each file group individually with proper ID tracking
	for _, fileGroup := range hlsData.FileGroups {
		// Set hash ID for file
		fileGroup.File.HashID = hlsData.Fileset.HashID

		// 1. Insert file → get fileID
		isSA := strings.HasSuffix(strings.ToUpper(hlsData.Fileset.ID), "SA")
		fileID, err := d.insertHLSFileTx(tx, fileGroup.File, isSA)
		if err != nil {
			return log.Error(d.ctx, 500, err, "Failed to insert HLS file")
		}

		// 2. Insert bandwidths for this file → collect bandwidthIDs
		bandwidthIDs := make([]int64, 0)
		for _, bandwidth := range fileGroup.Bandwidths {
			bandwidth.BibleFileID = fileID
			bandwidthID, err := d.insertHLSStreamBandwidthTx(tx, bandwidth)
			if err != nil {
				return log.Error(d.ctx, 500, err, "Failed to insert HLS stream bandwidth")
			}
			bandwidthIDs = append(bandwidthIDs, bandwidthID)
		}

		// 3. Insert bytes for this file → use correct bandwidthID
		for _, streamByte := range fileGroup.Bytes {
			if len(bandwidthIDs) > 0 {
				// For single bandwidth (current audio streams), use the first (and only) bandwidth
				// For future multi-bandwidth support, we'd need more sophisticated mapping
				streamByte.StreamBandwidthID = bandwidthIDs[0]
			}
			_, err := d.insertHLSStreamBytesTx(tx, streamByte)
			if err != nil {
				return log.Error(d.ctx, 500, err, "Failed to insert HLS stream bytes")
			}
		}
	}

	// For SA filesets, copy the stock_no tag from the corresponding DA fileset
	if isSA {
		// Convert SA fileset ID to DA fileset ID (replace "SA" with "DA")
		daFilesetID := strings.TrimSuffix(hlsData.Fileset.ID, "SA") + "DA"

		// Get the DA fileset's hash_id
		daHashID, status := d.SelectHashId(daFilesetID)
		if status != nil {
			return log.Error(d.ctx, 500, nil, "Failed to get DA fileset hash_id for stock_no tag copying")
		}

		// Copy the stock_no tag
		err = d.copyStockNoTagTx(tx, hlsData.Fileset.HashID, daHashID)
		if err != nil {
			return log.Error(d.ctx, 500, err, "Failed to copy stock_no tag from DA to SA fileset")
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to commit HLS data transaction")
	}

	return nil
}

// Transaction helper methods
func (d *DBPAdapter) insertHLSFilesetTx(tx *sql.Tx, fileset HLSFileset, isSA bool) (int64, error) {
	// Set content_loaded to 1 for SA filesets, 0 for others
	contentLoaded := 0
	if isSA {
		contentLoaded = 1
	}

	query := `INSERT INTO bible_filesets (id, set_type_code, set_size_code, mode_id, hash_id, asset_id, license_group_id, published_snm, content_loaded, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.Exec(query, fileset.ID, fileset.SetTypeCode, fileset.SetSizeCode, fileset.ModeID, fileset.HashID, fileset.AssetID, fileset.LicenseGroupID, fileset.PublishedSNM, contentLoaded, fileset.CreatedAt, fileset.UpdatedAt)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (d *DBPAdapter) insertHLSFileTx(tx *sql.Tx, file HLSFile, isSA bool) (int64, error) {
	var query string
	var result sql.Result
	var err error

	if isSA {
		// For SA filesets, set verse_start to "1", verse_end to NULL, and duration (rounded to int)
		query = `INSERT INTO bible_files (hash_id, book_id, chapter_start, verse_start, verse_end, file_name, file_size, duration, created_at, updated_at) 
				  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		result, err = tx.Exec(query, file.HashID, file.BookID, file.ChapterNum, "1", nil, file.FileName, file.FileSize, file.Duration, file.CreatedAt, file.UpdatedAt)
	} else {
		// For non-SA filesets, don't set verse_start/verse_end or duration
		query = `INSERT INTO bible_files (hash_id, book_id, chapter_start, file_name, file_size, created_at, updated_at) 
				  VALUES (?, ?, ?, ?, ?, ?, ?)`
		result, err = tx.Exec(query, file.HashID, file.BookID, file.ChapterNum, file.FileName, file.FileSize, file.CreatedAt, file.UpdatedAt)
	}

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// copyStockNoTagTx copies the stock_no tag from DA fileset to SA fileset
func (d *DBPAdapter) copyStockNoTagTx(tx *sql.Tx, saHashID, daHashID string) error {
	// Copy the stock_no tag from DA fileset to SA fileset
	query := `INSERT INTO bible_fileset_tags (hash_id, name, description, admin_only, notes, iso, language_id)
			  SELECT ?, name, description, admin_only, notes, iso, language_id
			  FROM bible_fileset_tags
			  WHERE hash_id = ? AND name = 'stock_no'`

	_, err := tx.Exec(query, saHashID, daHashID)
	return err
}

func (d *DBPAdapter) insertHLSStreamBandwidthTx(tx *sql.Tx, bandwidth HLSStreamBandwidth) (int64, error) {
	query := `INSERT INTO bible_file_stream_bandwidths (bible_file_id, file_name, bandwidth, resolution_width, resolution_height, codec, stream, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.Exec(query, bandwidth.BibleFileID, bandwidth.FileName, bandwidth.Bandwidth,
		bandwidth.ResolutionWidth, bandwidth.ResolutionHeight, bandwidth.Codec, bandwidth.Stream,
		bandwidth.CreatedAt, bandwidth.UpdatedAt)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (d *DBPAdapter) insertHLSStreamBytesTx(tx *sql.Tx, streamBytes HLSStreamBytes) (int64, error) {
	query := `INSERT INTO bible_file_stream_bytes (stream_bandwidth_id, runtime, bytes, offset, timestamp_id, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.Exec(query, streamBytes.StreamBandwidthID, streamBytes.Runtime,
		streamBytes.Bytes, streamBytes.Offset, streamBytes.TimestampID,
		streamBytes.CreatedAt, streamBytes.UpdatedAt)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
