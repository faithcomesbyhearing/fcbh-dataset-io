package update

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"math"
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

	// Check if we have both timestamps and HLS stanzas
	hasTimestamps := d.req.UpdateDBP.Timestamps != ""
	hasHLS := d.req.UpdateDBP.HLS != ""

	if hasTimestamps && hasHLS {
		// Both stanzas present - single transaction for both
		status = d.ProcessBothTimestampsAndHLS(chapters)
		if status != nil {
			return status
		}
	} else if hasTimestamps {
		// Only timestamps - single transaction for timestamps
		status = d.ProcessTimestampsOnly(chapters)
		if status != nil {
			return status
		}
	} else if hasHLS {
		// Only HLS - single transaction for HLS
		status = d.ProcessHLSOnly()
		if status != nil {
			return status
		}
	}

	return nil
}

func (d *UpdateTimestamps) ProcessBothTimestampsAndHLS(chapters []db.Script) *log.Status {
	log.Info(d.ctx, "Processing both timestamps and HLS in single transaction")

	// Remove all existing timestamps for the fileset once (not per chapter)
	status := d.dbpConn.RemoveTimestampsForFileset(d.req.UpdateDBP.Timestamps)
	if status != nil {
		return status
	}

	// Process timestamps first
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

			// Insert new timestamps (removal already done above)
			status = d.InsertTimestampsForFileset(d.req.UpdateDBP.Timestamps, ch.BookId, ch.ChapterNum, timestamps)
			if status != nil {
				return status
			}
		}
	}

	// Set timing estimation tag for timestamps fileset
	hashId, status := d.dbpConn.SelectHashId(d.req.UpdateDBP.Timestamps)
	if status != nil {
		return status
	}
	_, status = d.dbpConn.UpdateFilesetTimingEstTag(hashId, mmsAlignTimingEstErr)
	if status != nil {
		return status
	}

	// Now process HLS with the newly inserted timestamps
	status = d.ProcessHLS(d.req.UpdateDBP.HLS, d.req.BibleId)
	if status != nil {
		return status
	}

	log.Info(d.ctx, "Both timestamps and HLS updated successfully")
	return nil
}

func (d *UpdateTimestamps) ProcessTimestampsOnly(chapters []db.Script) *log.Status {
	log.Info(d.ctx, "Processing timestamps only")

	// Remove all existing timestamps for the fileset once (not per chapter)
	status := d.dbpConn.RemoveTimestampsForFileset(d.req.UpdateDBP.Timestamps)
	if status != nil {
		return status
	}

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

			// Insert new timestamps (removal already done above)
			status = d.InsertTimestampsForFileset(d.req.UpdateDBP.Timestamps, ch.BookId, ch.ChapterNum, timestamps)
			if status != nil {
				return status
			}
		}
	}

	// Set timing estimation tag for timestamps fileset
	hashId, status := d.dbpConn.SelectHashId(d.req.UpdateDBP.Timestamps)
	if status != nil {
		return status
	}
	_, status = d.dbpConn.UpdateFilesetTimingEstTag(hashId, mmsAlignTimingEstErr)
	if status != nil {
		return status
	}

	log.Info(d.ctx, "Timestamps updated successfully")
	return nil
}

func (d *UpdateTimestamps) ProcessHLSOnly() *log.Status {
	log.Info(d.ctx, "Processing HLS only")

	// Process HLS using the existing ProcessHLS function
	status := d.ProcessHLS(d.req.UpdateDBP.HLS, d.req.BibleId)
	if status != nil {
		return status
	}

	log.Info(d.ctx, "HLS updated successfully")
	return nil
}

func (d *UpdateTimestamps) UpdateFileset(filesetId string, bookId string, chapterNum int, timestamps []Timestamp) *log.Status {
	var status *log.Status
	var hashId string
	hashId, status = d.dbpConn.SelectHashId(filesetId)
	if status != nil {
		return status
	}
	var bibleFileId int64
	bibleFileId, _, status = d.dbpConn.SelectFileId(hashId, bookId, chapterNum)
	if status != nil {
		return status // what is the correct response for not found
	}
	if bibleFileId > 0 {
		var dbpTimestamps []Timestamp
		dbpTimestamps, status = d.dbpConn.SelectTimestamps(bibleFileId)
		if status != nil {
			return status
		}
		if len(dbpTimestamps) > 0 {
			timestamps = MergeTimestamps(timestamps, dbpTimestamps)
		}
		// Update existing timestamps
		_, status = d.dbpConn.UpdateTimestamps(timestamps)
		if status != nil {
			return status
		}

		// Insert new timestamps
		_, _, status = d.dbpConn.InsertTimestamps(bibleFileId, timestamps)
		if status != nil {
			return status
		}

		// Only process HLS segments and timing estimation if not just doing timestamps
		// TODO: Add HLS stanza check here when HLS processing is needed
		// For now, skip HLS processing when only timestamps are requested
		log.Info(d.ctx, "Skipping HLS processing - timestamps only mode")
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

	// Get all chapters that have timestamps from MySQL (not SQLite)
	chapters, status := d.dbpConn.SelectBookChapterFromDBP(timestampsFilesetID)
	if status != nil {
		return status
	}

	// Get mode_id and license info from the source timestamps fileset
	modeID, licenseGroupID, publishedSNM, status := d.dbpConn.SelectFilesetLicenseInfo(timestampsFilesetID)
	if status != nil {
		return status
	}

	// Collect all HLS data
	var hlsData HLSData
	now := time.Now().Format("2006-01-02 15:04:05")
	hlsData.Fileset = HLSFileset{
		ID:             hlsFilesetID,
		SetTypeCode:    "audio_stream",                               // Default to audio_stream, could be made configurable
		SetSizeCode:    "NT",                                         // Default to NT, could be made configurable
		ModeID:         modeID,                                       // Copy from source timestamps fileset
		HashID:         generateHashID(hlsFilesetID, "audio_stream"), // Generate a unique hash ID
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
					CreatedAt:  now,
					UpdatedAt:  now,
				},
				Bandwidths: fileData.Bandwidths,
				Bytes:      fileData.Bytes,
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

func (d *UpdateTimestamps) ReplaceTimestampsForFileset(filesetID, bookID string, chapterNum int, timestamps []Timestamp) *log.Status {
	// Remove existing timestamps for this fileset
	status := d.dbpConn.RemoveTimestampsForFileset(filesetID)
	if status != nil {
		return status
	}

	// Insert new timestamps
	hashID, status := d.dbpConn.SelectHashId(filesetID)
	if status != nil {
		return status
	}

	bibleFileID, _, status := d.dbpConn.SelectFileId(hashID, bookID, chapterNum)
	if status != nil {
		return status
	}

	if bibleFileID > 0 {
		_, _, status = d.dbpConn.InsertTimestamps(bibleFileID, timestamps)
		if status != nil {
			return status
		}
	}

	return nil
}

func (d *UpdateTimestamps) InsertTimestampsForFileset(filesetID, bookID string, chapterNum int, timestamps []Timestamp) *log.Status {
	// Insert new timestamps (removal already done outside this function)
	hashID, status := d.dbpConn.SelectHashId(filesetID)
	if status != nil {
		return status
	}

	bibleFileID, _, status := d.dbpConn.SelectFileId(hashID, bookID, chapterNum)
	if status != nil {
		return status
	}

	if bibleFileID > 0 {
		// Check if we need to add a verse 0 entry
		var enhancedTimestamps []Timestamp
		if len(timestamps) > 0 && timestamps[0].BeginTS > 0.0 {
			// First verse starts after 0, add verse 0 entry to capture intro/header
			enhancedTimestamps = make([]Timestamp, 0, len(timestamps)+1)
			verse0Entry := Timestamp{
				VerseStr:    "0",
				VerseEnd:    sql.NullString{Valid: false},
				BeginTS:     0.0,
				EndTS:       timestamps[0].BeginTS,
				VerseSeq:    0,
				TimestampId: 0, // Will be set by InsertTimestamps
			}
			enhancedTimestamps = append(enhancedTimestamps, verse0Entry)
			enhancedTimestamps = append(enhancedTimestamps, timestamps...)
		} else {
			// First verse starts at 0 or no timestamps, use as-is
			enhancedTimestamps = timestamps
		}

		_, _, status = d.dbpConn.InsertTimestamps(bibleFileID, enhancedTimestamps)
		if status != nil {
			return status
		}
	}

	return nil
}

func generateHashID(filesetID, setTypeCode string) string {
	// Generate hash_id using the same method as DBP: MD5(filesetID + bucket + setTypeCode)[:12]
	// bucket is typically "dbp-prod"
	bucket := "dbp-prod"

	// Create MD5 hash
	hash := md5.Sum([]byte(filesetID + bucket + setTypeCode))

	// Convert to hex string and truncate to 12 characters
	return fmt.Sprintf("%x", hash)[:12]
}
