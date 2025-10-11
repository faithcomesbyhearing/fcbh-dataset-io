package update

import (
	"context"
	"database/sql"
	"errors"
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
	query := `SELECT hash_id FROM bible_filesets WHERE asset_id = 'dbp-prod'
		AND set_type_code IN ('audio', 'audio_drama', 'audio_stream', 'audio_drama_stream') 
		AND id = ?`
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

func (d *DBPAdapter) UpdateTimestamps(timestamps []Timestamp) (int, *log.Status) {
	var rowCount int
	var mustUpdate int
	for _, rec := range timestamps {
		if rec.TimestampId > 0 {
			mustUpdate++
		}
	}
	if mustUpdate > 0 {
		query := `UPDATE bible_file_timestamps SET timestamp = ?, timestamp_end = ? 
			WHERE id = ?`
		tx, err := d.conn.Begin()
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		stmt, err := tx.Prepare(query)
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		defer stmt.Close()
		var result sql.Result
		var count int64
		for _, rec := range timestamps {
			if rec.TimestampId > 0 {
				result, err = stmt.Exec(rec.BeginTS, rec.EndTS, rec.TimestampId)
				if err != nil {
					return rowCount, log.Error(d.ctx, 500, err, query)
				}
				count, err = result.RowsAffected()
				if err != nil {
					return rowCount, log.Error(d.ctx, 500, err, query)
				}
				rowCount += int(count)
			}
		}
		err = tx.Commit()
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		if rowCount != mustUpdate {
			return rowCount, log.ErrorNoErr(d.ctx, 500, "Row count expected:",
				mustUpdate, "Actual Count:", rowCount, query)
		}
	}
	return rowCount, nil
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

func (d *DBPAdapter) UpdateSegments(segments []Timestamp) (int, *log.Status) {
	var rowCount int
	query := `UPDATE bible_file_stream_bytes SET runtime = ?, offset = ?, bytes = ?
		WHERE timestamp_id = ?`
	tx, err := d.conn.Begin()
	if err != nil {
		return rowCount, log.Error(d.ctx, 500, err, query)
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		return rowCount, log.Error(d.ctx, 500, err, query)
	}
	defer stmt.Close()
	var result sql.Result
	var count int64
	for _, rec := range segments {
		result, err = stmt.Exec(rec.Duration, rec.Position, rec.NumBytes, rec.TimestampId)
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, `Error while inserting dbp timestamp.`)
		}
		count, err = result.RowsAffected()
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		rowCount += int(count)
	}
	err = tx.Commit()
	if err != nil {
		return rowCount, log.Error(d.ctx, 500, err, query)
	}
	return rowCount, nil
}

func (d *DBPAdapter) UpdateFilesetTimingEstTag(hashId string, timingEstErr string) (int, *log.Status) {
	var rowCount int
	query := `SELECT description FROM bible_fileset_tags WHERE hash_id = ? AND name = 'timing_est_err'`
	var currEstErr string
	err := d.conn.QueryRow(query, hashId).Scan(&currEstErr)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, log.Error(d.ctx, 500, err, query)
	}
	var result sql.Result
	var count int64
	if errors.Is(err, sql.ErrNoRows) {
		query = `INSERT INTO bible_fileset_tags (hash_id, name, description, admin_only, iso, language_id)
		VALUES (?, 'timing_est_err', ?, 0, 'eng', 6414)`
		result, err = d.conn.Exec(query, hashId, timingEstErr)
		if err != nil {
			return 0, log.Error(d.ctx, 500, err, query)
		}
		count, err = result.RowsAffected()
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		rowCount = int(count)
	} else if currEstErr != timingEstErr {
		query = `UPDATE bible_fileset_tags SET description = ? WHERE hash_id = ? AND name = 'timing_est_err'`
		result, err = d.conn.Exec(query, timingEstErr, hashId)
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		count, err = result.RowsAffected()
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		rowCount += int(count)
		if rowCount != 1 {
			return rowCount, log.ErrorNoErr(d.ctx, 500,
				"Row count expected:", 1, "Actual Count:", rowCount, query)
		}
	}
	return rowCount, nil
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

func (d *DBPAdapter) InsertHLSFileset(fileset HLSFileset) (int64, *log.Status) {
	query := `INSERT INTO bible_filesets (id, set_type_code, set_size_code, hash_id, asset_id, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, 'dbp-prod', ?, ?)`

	result, err := d.conn.Exec(query, fileset.ID, fileset.SetTypeCode, fileset.SetSizeCode, fileset.HashID, fileset.CreatedAt, fileset.UpdatedAt)
	if err != nil {
		return 0, log.Error(d.ctx, 500, err, query)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, log.Error(d.ctx, 500, err, "Failed to get last insert ID")
	}

	return id, nil
}

func (d *DBPAdapter) InsertHLSFile(file HLSFile, isSA bool) (int64, *log.Status) {
	var query string
	var result sql.Result
	var err error

	if isSA {
		// For SA filesets, set verse_start to "1" and verse_end to NULL
		query = `INSERT INTO bible_files (hash_id, book_id, chapter_start, verse_start, verse_end, file_name, file_size, created_at, updated_at) 
				  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		result, err = d.conn.Exec(query, file.HashID, file.BookID, file.ChapterNum, "1", nil, file.FileName, file.FileSize, file.CreatedAt, file.UpdatedAt)
	} else {
		// For non-SA filesets, don't set verse_start/verse_end
		query = `INSERT INTO bible_files (hash_id, book_id, chapter_start, file_name, file_size, created_at, updated_at) 
				  VALUES (?, ?, ?, ?, ?, ?, ?)`
		result, err = d.conn.Exec(query, file.HashID, file.BookID, file.ChapterNum, file.FileName, file.FileSize, file.CreatedAt, file.UpdatedAt)
	}

	if err != nil {
		return 0, log.Error(d.ctx, 500, err, query)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, log.Error(d.ctx, 500, err, "Failed to get last insert ID")
	}

	return id, nil
}

func (d *DBPAdapter) InsertHLSStreamBandwidth(bandwidth HLSStreamBandwidth) (int64, *log.Status) {
	query := `INSERT INTO bible_file_stream_bandwidths (bible_file_id, file_name, bandwidth, resolution_width, resolution_height, codec, stream, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := d.conn.Exec(query, bandwidth.BibleFileID, bandwidth.FileName, bandwidth.Bandwidth,
		bandwidth.ResolutionWidth, bandwidth.ResolutionHeight, bandwidth.Codec, bandwidth.Stream,
		bandwidth.CreatedAt, bandwidth.UpdatedAt)
	if err != nil {
		return 0, log.Error(d.ctx, 500, err, query)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, log.Error(d.ctx, 500, err, "Failed to get last insert ID")
	}

	return id, nil
}

func (d *DBPAdapter) InsertHLSStreamBytes(streamBytes HLSStreamBytes) (int64, *log.Status) {
	query := `INSERT INTO bible_file_stream_bytes (stream_bandwidth_id, runtime, bytes, offset, timestamp_id, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := d.conn.Exec(query, streamBytes.StreamBandwidthID, streamBytes.Runtime,
		streamBytes.Bytes, streamBytes.Offset, streamBytes.TimestampID,
		streamBytes.CreatedAt, streamBytes.UpdatedAt)
	if err != nil {
		return 0, log.Error(d.ctx, 500, err, query)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, log.Error(d.ctx, 500, err, "Failed to get last insert ID")
	}

	return id, nil
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

func (d *DBPAdapter) RemoveTimestampsForFileset(filesetID string) *log.Status {
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

	// Delete timestamps for all files in this fileset
	_, err = tx.Exec(`DELETE FROM bible_file_timestamps WHERE bible_file_id IN (SELECT id FROM bible_files WHERE hash_id = ?)`, hashID)
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to delete timestamps for fileset: "+filesetID)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to commit timestamp removal transaction")
	}

	return nil
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
	_, err = d.insertHLSFilesetTx(tx, hlsData.Fileset)
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to insert HLS fileset")
	}

	// Process each file group individually with proper ID tracking
	for _, fileGroup := range hlsData.FileGroups {
		// Set hash ID for file
		fileGroup.File.HashID = hlsData.Fileset.HashID

		// 1. Insert file → get fileID
		isSA := d.isSAFileset(hlsData.Fileset.ID)
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

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return log.Error(d.ctx, 500, err, "Failed to commit HLS data transaction")
	}

	return nil
}

// Transaction helper methods
func (d *DBPAdapter) insertHLSFilesetTx(tx *sql.Tx, fileset HLSFileset) (int64, error) {
	query := `INSERT INTO bible_filesets (id, set_type_code, set_size_code, mode_id, hash_id, asset_id, license_group_id, published_snm, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, 'dbp-prod', ?, ?, ?, ?)`

	result, err := tx.Exec(query, fileset.ID, fileset.SetTypeCode, fileset.SetSizeCode, fileset.ModeID, fileset.HashID, fileset.LicenseGroupID, fileset.PublishedSNM, fileset.CreatedAt, fileset.UpdatedAt)
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

// isSAFileset checks if a fileset ID represents an SA (Single Audio) fileset
func (d *DBPAdapter) isSAFileset(filesetID string) bool {
	// SA filesets typically end with "SA" (e.g., ENGNIVN1SA)
	return strings.HasSuffix(strings.ToUpper(filesetID), "SA")
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
