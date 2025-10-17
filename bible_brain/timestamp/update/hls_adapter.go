package update

import (
	"database/sql"
	"strings"

	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

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
