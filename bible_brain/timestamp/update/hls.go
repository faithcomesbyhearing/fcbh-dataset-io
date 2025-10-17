package update

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

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
