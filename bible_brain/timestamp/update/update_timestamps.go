package update

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

const (
	mmsAlignTimingEstErr = "mms_align"
)

type UpdateTimestamps struct {
	ctx     context.Context
	conn    db.DBAdapter
	dbpConn DBPAdapter
}

func (d *UpdateTimestamps) Process(filesetId string) *log.Status {
	var status *log.Status
	status = d.dbpConn.DeleteTimestamps(filesetId)
	if status != nil {
		return status
	}
	var hashId string
	hashId, status = d.dbpConn.SelectHashId(filesetId)
	if status != nil {
		return status
	}
	status = d.dbpConn.UpdateFilesetTimingEstTag(hashId, mmsAlignTimingEstErr)
	if status != nil {
		return status
	}
	var books []string
	books = append(books, db.BookOT...)
	books = append(books, db.BookNT...)
	for _, book := range books {
		lastChapter, _ := db.BookChapterMap[book]
		for chap := 1; chap <= lastChapter; chap++ {
			var timestamps []db.Audio
			timestamps, status = d.conn.SelectFAScriptTimestamps(book, chap)
			if status != nil {
				return status
			}
			if len(timestamps) > 0 {
				var bibleFileId int64
				bibleFileId, status = d.dbpConn.SelectFileId(filesetId,
					timestamps[0].BookId, timestamps[0].ChapterNum)
				if status != nil {
					return status
				}
				status = d.dbpConn.UpdateTimestamps(bibleFileId, timestamps)
				if status != nil {
					return status
				}
			}
		}
	}
	return status
}
