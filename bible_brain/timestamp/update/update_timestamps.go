package update

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"math"
	"os"
	"path/filepath"
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
	bibleId := ident.BibleId
	filesetId := ident.AudioNTId // base this on bookId
	if filesetId == "" {
		filesetId = ident.AudioOTId
	}
	log.Info(d.ctx, "Processing:", filesetId)
	var chapters []db.Script
	chapters, status = d.conn.SelectBookChapter()
	if status != nil {
		return status
	}
	for _, ch := range chapters {
		var timestamps []Timestamp
		timestamps, status = d.SelectFATimestamps(ch.BookId, ch.ChapterNum)
		if status != nil {
			return status
		}
		if len(timestamps) > 0 {
			directory := os.Getenv("FCBH_DATASET_FILES")
			audioPath := filepath.Join(directory, bibleId, filesetId, timestamps[0].AudioFile)
			timestamps, status = ComputeBytes(d.ctx, audioPath, timestamps)
			if status != nil {
				return status
			}
			for i := range timestamps {
				timestamps[i].BeginTS = math.Round(timestamps[i].BeginTS*100.0) / 100.0
				timestamps[i].EndTS = math.Round(timestamps[i].EndTS*100.0) / 100.0
			}
			for _, subFilesetId := range d.req.UpdateDBP.Timestamps {
				log.Info(d.ctx, "Updating:", subFilesetId, ch.BookId, ch.ChapterNum)
				status = d.UpdateFileset(subFilesetId, ch.BookId, ch.ChapterNum, timestamps)
				if status != nil {
					return status
				}
			}
		}
	}
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
		_, status = d.dbpConn.UpdateTimestamps(timestamps)
		if status != nil {
			return status
		}
		timestamps, _, status = d.dbpConn.InsertTimestamps(bibleFileId, timestamps)
		if status != nil {
			return status
		}
		_, status = d.dbpConn.UpdateSegments(timestamps)
		if status != nil {
			return status
		}
		_, status = d.dbpConn.UpdateFilesetTimingEstTag(hashId, mmsAlignTimingEstErr)
		if status != nil {
			return status
		}
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
