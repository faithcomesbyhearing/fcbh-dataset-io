package update

import (
	"context"
	"math"

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
			timestamps, status := d.SelectTimestampsFromSQLite(ch.BookId, ch.ChapterNum)
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

func (d *UpdateTimestamps) SelectTimestampsFromSQLite(bookId string, chapter int) ([]Timestamp, *log.Status) {
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
