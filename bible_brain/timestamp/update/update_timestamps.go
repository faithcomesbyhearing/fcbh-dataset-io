package update

import (
	"context"
	"crypto/md5"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

const (
	mmsAlignTimingEstErr = "mms_align"
)

type UpdateTimestamps struct {
	ctx     context.Context
	req     request.Request
	conn    db.DBAdapter
	dbpConn DBPAdapter
}

func NewUpdateTimestamps(ctx context.Context, req request.Request, conn db.DBAdapter) UpdateTimestamps {
	var u UpdateTimestamps
	u.ctx = ctx
	u.req = req // This could be only the dbp_update timestamps
	u.conn = conn
	return u
}

func (d *UpdateTimestamps) Process() *log.Status {
	var status *log.Status
	d.dbpConn, status = NewDBPAdapter(d.ctx)
	if status != nil {
		return status
	}
	defer d.dbpConn.Close()
	var ident db.Ident
	ident, status = d.conn.SelectIdent()
	if status != nil {
		return status
	}
	filesetId := ident.AudioNTId // base this on bookId
	if filesetId == "" {
		filesetId = ident.AudioOTId
	}
	log.Info(d.ctx, "DBP Update Processing:", filesetId)
	var chapters []db.Script
	chapters, status = d.conn.SelectBookChapter()
	if status != nil {
		return status
	}

	// Process timestamps if specified
	if d.req.UpdateDBP.Timestamps != "" {
		// Collect all timestamps for all chapters
		timestampsData := make(map[string]map[int][]Timestamp)
		for _, ch := range chapters {
			var timestamps []Timestamp
			timestamps, status := d.SelectFATimestamps(ch.BookId, ch.ChapterNum)
			if status != nil {
				return status
			}
			if len(timestamps) > 0 {
				// Round timestamps to 3 decimal places
				for i := range timestamps {
					timestamps[i].BeginTS = math.Round(timestamps[i].BeginTS*1000.0) / 1000.0
					timestamps[i].EndTS = math.Round(timestamps[i].EndTS*1000.0) / 1000.0
				}

				// Add to map
				if timestampsData[ch.BookId] == nil {
					timestampsData[ch.BookId] = make(map[int][]Timestamp)
				}
				timestampsData[ch.BookId][ch.ChapterNum] = timestamps
			}
		}

		// Process timestamps in a single transaction (removes SA files, removes/inserts DA timestamps, updates tag)
		status = d.dbpConn.ProcessTimestampsWithTag(d.req.UpdateDBP.Timestamps, mmsAlignTimingEstErr, chapters, timestampsData)
		if status != nil {
			return status
		}
		log.Info(d.ctx, "Timestamps updated successfully")
	}

	// Process HLS if specified
	if d.req.UpdateDBP.HLS != "" {
		status = d.ProcessHLS(d.req.UpdateDBP.HLS, d.req.BibleId)
		if status != nil {
			return status
		}
		log.Info(d.ctx, "HLS updated successfully")
	}

	return nil
}

func (d *UpdateTimestamps) SelectFATimestamps(bookId string, chapter int) ([]Timestamp, *log.Status) {
	var result []Timestamp
	datasetTS, status := d.conn.SelectFAScriptTimestamps(bookId, chapter)
	if status != nil {
		return result, status
	}
	for i, db := range datasetTS {
		var t Timestamp
		t.VerseStr = db.VerseStr
		if db.VerseEnd == "" {
			t.VerseEnd.Valid = false
		} else {
			t.VerseEnd.String = db.VerseEnd
			t.VerseEnd.Valid = true
		}
		t.VerseSeq = i + 1
		t.BeginTS = db.BeginTS
		t.EndTS = db.EndTS
		t.AudioFile = db.AudioFile
		result = append(result, t)
	}
	return result, nil
}

func MergeTimestamps(timestamps []Timestamp, dbpTimestamps []Timestamp) []Timestamp {
	var dbpMap = make(map[string]Timestamp)
	for _, dbp := range dbpTimestamps {
		dbpMap[dbp.VerseStr] = dbp
	}
	for i, ts := range timestamps {
		dbp, ok := dbpMap[ts.VerseStr]
		if ok {
			timestamps[i].TimestampId = dbp.TimestampId
		}
	}
	return timestamps
}

func (d *UpdateTimestamps) ProcessHLS(hlsFilesetID, bibleID string) *log.Status {
	// Initialize DBP connection if not already done
	if d.dbpConn.conn == nil {
		var status *log.Status
		d.dbpConn, status = NewDBPAdapter(d.ctx)
		if status != nil {
			return status
		}
	}
	defer d.dbpConn.Close()

	// Get the timestamps fileset ID from the request
	timestampsFilesetID := d.req.UpdateDBP.Timestamps
	if timestampsFilesetID == "" {
		return log.ErrorNoErr(d.ctx, 400, "Timestamps fileset ID required for HLS processing")
	}

	// Create HLS processor
	processor := NewLocalHLSProcessor(d.ctx, bibleID, timestampsFilesetID)

	// Get chapters from SQLite (only process books that have timestamps in the dataset)
	chapters, status := d.conn.SelectBookChapter()
	if status != nil {
		return status
	}

	// Get mode_id and license info from the source timestamps fileset
	modeID, licenseGroupID, publishedSNM, status := d.dbpConn.SelectFilesetLicenseInfo(timestampsFilesetID)
	if status != nil {
		return status
	}

	// Get asset_id for hash generation
	// For SA filesets, use the parent DA fileset's asset_id
	var assetID string
	if strings.HasSuffix(strings.ToUpper(hlsFilesetID), "SA") {
		// Convert SA fileset ID to DA fileset ID (replace "SA" with "DA")
		daFilesetID := strings.TrimSuffix(hlsFilesetID, "SA") + "DA"
		assetID, status = d.dbpConn.SelectAssetId(daFilesetID)
		if status != nil {
			return log.Error(d.ctx, 500, nil, "Failed to get DA fileset asset_id for SA fileset: "+hlsFilesetID)
		}
	} else {
		// For non-SA filesets, use the timestamps fileset's asset_id
		assetID, status = d.dbpConn.SelectAssetId(timestampsFilesetID)
		if status != nil {
			return log.Error(d.ctx, 500, nil, "Failed to get timestamps fileset asset_id: "+timestampsFilesetID)
		}
	}

	// Collect all HLS data
	var hlsData HLSData
	now := time.Now().Format("2006-01-02 15:04:05")
	hlsData.Fileset = HLSFileset{
		ID:             hlsFilesetID,
		SetTypeCode:    "audio_stream",                                        // Default to audio_stream, could be made configurable
		SetSizeCode:    "NT",                                                  // Default to NT, could be made configurable
		ModeID:         modeID,                                                // Copy from source timestamps fileset
		HashID:         generateHashID(hlsFilesetID, "audio_stream", assetID), // Generate a unique hash ID using asset_id as bucket
		AssetID:        assetID,                                               // Use the asset_id from parent DA fileset or timestamps fileset
		BibleID:        bibleID,
		LicenseGroupID: licenseGroupID, // Copy from source timestamps fileset
		PublishedSNM:   publishedSNM,   // Copy from source timestamps fileset
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Process each chapter and create file groups
	for _, ch := range chapters {
		// Get timestamps for this chapter from MySQL (not SQLite)
		timestamps, status := d.dbpConn.SelectFATimestampsFromDBP(ch.BookId, ch.ChapterNum, timestampsFilesetID)
		if status != nil {
			return status
		}

		if len(timestamps) > 0 {
			// Find the audio file for this chapter
			audioFile := timestamps[0].AudioFile
			if audioFile == "" {
				log.Info(d.ctx, "No audio file found for chapter:", ch.BookId, ch.ChapterNum)
				continue
			}

			// Process the file with HLS processor
			fileData, err := processor.ProcessFile(audioFile, timestamps)
			if err != nil {
				return log.Error(d.ctx, 500, err, "Failed to process HLS file: "+audioFile)
			}

			// Create file group for this chapter
			fileGroup := HLSFileGroup{
				File: HLSFile{
					BookID:     ch.BookId,
					ChapterNum: ch.ChapterNum,
					FileName:   fileData.File.FileName,
					FileSize:   fileData.File.FileSize,
					Duration:   fileData.File.Duration,
					CreatedAt:  now,
					UpdatedAt:  now,
				},
				Bandwidths: fileData.Bandwidths,
				Bytes:      fileData.Bytes,
			}

			// Special handling for SA filesets: set verse_start to 1
			if strings.HasSuffix(strings.ToUpper(hlsFilesetID), "SA") {
				// For SA filesets, we need to ensure verse_start is always 1
				// This is handled in the database insertion logic
				log.Info(d.ctx, "Creating HLS for:", hlsFilesetID, " ", ch.BookId, " ", ch.ChapterNum)
			}

			// Add file group to HLS data
			hlsData.FileGroups = append(hlsData.FileGroups, fileGroup)
		}
	}

	// Insert all HLS data atomically
	status = d.dbpConn.InsertHLSData(hlsData)
	if status != nil {
		return status
	}

	log.Info(d.ctx, "Successfully processed HLS for fileset:", hlsFilesetID)
	return nil
}

func generateHashID(filesetID, setTypeCode, bucket string) string {
	// Generate hash_id using the same method as DBP: MD5(filesetID + bucket + setTypeCode)[:12]
	// bucket is typically "dbp-prod" or the asset_id from parent DA fileset

	// Create MD5 hash
	hash := md5.Sum([]byte(filesetID + bucket + setTypeCode))

	// Convert to hex string and truncate to 12 characters
	return fmt.Sprintf("%x", hash)[:12]
}
