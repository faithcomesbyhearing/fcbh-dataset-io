package update

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"strings"

	//"github.com/faithcomesbyhearing/fcbh-dataset-io/fetch"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
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

func NewUpdateTimestamps(ctx context.Context, req request.Request, conn db.DBAdapter, dbp DBPAdapter) UpdateTimestamps {
	var u UpdateTimestamps
	u.ctx = ctx
	u.req = req // This could be only the dbp_update timestamps
	u.conn = conn
	u.dbpConn = dbp
	return u
}

func (d *UpdateTimestamps) Process() *log.Status {
	directory := os.Getenv("FCBH_DATASET_FILES")
	for _, filesetId := range d.req.UpdateDBP.Timestamps {
		status := d.UpdateFileset(filesetId, directory)
		if status != nil {
			return status
		}
	}
	return nil
}

func (d *UpdateTimestamps) UpdateFileset(filesetId string, directory string) *log.Status {
	var status *log.Status
	var hashId string
	hashId, status = d.dbpConn.SelectHashId(filesetId)
	if status != nil {
		return status
	}
	var books []string
	books = append(books, db.BookOT...)
	books = append(books, db.BookNT...)
	for _, book := range books {
		lastChapter, _ := db.BookChapterMap[book]
		for chap := 1; chap <= lastChapter; chap++ {
			var timestamps []Timestamp
			timestamps, status = d.SelectTimestamps(book, chap)
			if len(timestamps) > 0 {
				var bibleFileId int64
				var audioFile string
				bibleFileId, audioFile, status = d.dbpConn.SelectFileId(hashId, book, chap)
				if status != nil {
					return status // what is the correct response for not found
				}
				if bibleFileId > 0 {
					var dbpTimestamps []Timestamp
					dbpTimestamps, status = d.dbpConn.SelectTimestamps(bibleFileId)
					if status != nil {
						return status
					}
					timestamps = MergeTimestamps(timestamps, dbpTimestamps)
					_, status = d.dbpConn.UpdateTimestamps(timestamps)
					if status != nil {
						return status
					}
					timestamps, _, status = d.dbpConn.InsertTimestamps(bibleFileId, timestamps)
					if status != nil {
						return status
					}
					audioPath := filepath.Join(directory, audioFile)
					timestamps, status = ComputeBytes(d.ctx, audioPath, timestamps)
					if status != nil {
						return status
					}
					status = d.dbpConn.UpdateSegments(timestamps)
					if status != nil {
						return status
					}
				}
			}
		}
	}
	status = d.dbpConn.UpdateFilesetTimingEstTag(hashId, mmsAlignTimingEstErr)
	if status != nil {
		return status
	}
	return nil
}

func (d *UpdateTimestamps) SelectTimestamps(bookId string, chapter int) ([]Timestamp, *log.Status) {
	var result []Timestamp
	datasetTS, status := d.conn.SelectFAScriptTimestamps(bookId, chapter)
	if status != nil {
		return result, status
	}
	for i, db := range datasetTS {
		var t Timestamp
		parts := strings.FieldsFunc(db.VerseStr, func(r rune) bool {
			return r == '-' || r == ','
		})
		if len(parts) > 0 {
			t.VerseStr = parts[0]
		}
		if len(parts) > 1 {
			t.VerseEnd.String = parts[len(parts)-1]
			t.VerseEnd.Valid = true
		} else if db.VerseEnd == "" {
			t.VerseEnd.Valid = false
		} else {
			t.VerseEnd.String = db.VerseEnd
			t.VerseEnd.Valid = true
		}
		t.VerseSeq = i + 1
		t.BeginTS = db.BeginTS
		t.EndTS = db.EndTS
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
