package read

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"strconv"
)

/**
This class uses DBPTextReader to get the Script records
But then it gets the mt and h headings from USX files,
And add this to the DBPText data, and stores the combined data.
*/

//Could this all be done is SQL?
//Select from each source 1 book at a time, and one chapter at a time

type DBPTextEditReader struct {
	ctx       context.Context
	bibleId   string
	conn      db.DBAdapter
	req       request.Request // This is not needed
	testament request.Testament
}

func NewDBPTextEditReader(conn db.DBAdapter, req request.Request) DBPTextEditReader {
	var d DBPTextEditReader
	d.ctx = conn.Ctx
	d.bibleId = req.BibleId
	d.conn = conn
	d.req = req
	d.testament = req.Testament
	return d
}

func (d *DBPTextEditReader) Process() *log.Status {
	var status *log.Status
	testament := d.req.Testament
	var usxDB db.DBAdapter
	usxDB, status = d.createUSXEDITText(testament)
	if status != nil {
		return status
	}
	titleMap, chapMap, status := d.readUSXHeadings(usxDB) //testament)
	if status != nil {
		return status
	}
	var textDB db.DBAdapter
	textDB, status = d.createDBPText(testament)
	if status != nil {
		return status
	}
	records, status := d.combineHeadings(textDB, titleMap, chapMap)
	if status != nil {
		return status
	}
	status = d.conn.InsertScripts(records)
	return status
}

func (d *DBPTextEditReader) createDBPText(testament request.Testament) (db.DBAdapter, *log.Status) {
	var database db.DBAdapter
	var status *log.Status
	var otMediaId string
	var ntMediaId string
	if testament.OT || len(testament.OTBooks) > 0 {
		otMediaId = d.bibleId + `O_ET`
	}
	if testament.NT || len(testament.NTBooks) > 0 {
		ntMediaId = d.bibleId + `N_ET`
	}
	files, status := input.DBPDirectory(d.ctx, d.bibleId, request.TextPlainEdit, otMediaId,
		ntMediaId)
	if status != nil {
		return database, status
	}
	database = db.NewDBAdapter(d.ctx, ":memory:")
	textAdapter := NewDBPTextReader(database, d.req.Testament)
	status = textAdapter.ProcessFiles(files)
	return database, status
}

func (d *DBPTextEditReader) createUSXEDITText(testament request.Testament) (db.DBAdapter, *log.Status) {
	var database db.DBAdapter
	var status *log.Status
	files, status := input.DBPDirectory(d.ctx, d.bibleId, request.TextUSXEdit, d.bibleId+`O_ET-usx`,
		d.bibleId+`N_ET-usx`)
	if status != nil {
		return database, status
	}
	database = db.NewDBAdapter(d.ctx, ":memory:")
	usx := NewUSXParser(database)
	status = usx.ProcessFiles(files)
	return database, status
}

func (d *DBPTextEditReader) readUSXHeadings(conn4 db.DBAdapter) (map[string]db.Script, map[string]db.Script, *log.Status) {
	var bookTitle = make(map[string]db.Script)
	var chapTitle = make(map[string]db.Script)
	var status *log.Status
	var records []db.Script
	records, status = conn4.SelectScriptHeadings()
	conn4.Close()
	if status == nil {
		for _, rec := range records {
			if rec.UsfmStyle == `para.h` {
				key := rec.BookId + `:` + strconv.Itoa(rec.ChapterNum)
				chapTitle[key] = rec
			} else {
				bookTitle[rec.BookId] = rec
			}
		}
	}
	return bookTitle, chapTitle, status
}

func (d *DBPTextEditReader) combineHeadings(conn db.DBAdapter, bookTitle map[string]db.Script,
	chapTitle map[string]db.Script) ([]db.Script, *log.Status) {
	var results = make([]db.Script, 0, 5000)
	var lastBookId = ``
	var lastChapter = -1
	var scriptNum = 0
	var records, status = conn.SelectScripts()
	conn.Close()
	if status != nil {
		return results, status
	}
	for _, rec := range records {
		if d.testament.HasOT(rec.BookId) || d.testament.HasNT(rec.BookId) {
			if lastBookId != rec.BookId {
				lastBookId = rec.BookId
				lastChapter = rec.ChapterNum // skip chapter heading on chapter 1
				var inp = rec
				inp.VerseStr = `0`
				inp.VerseNum = 0
				scriptNum = 1
				inp.ScriptNum = strconv.Itoa(scriptNum)
				titleRec := bookTitle[rec.BookId]
				inp.UsfmStyle = titleRec.UsfmStyle
				inp.ScriptTexts = []string{titleRec.ScriptText}
				results = append(results, inp)
			} else if lastChapter != rec.ChapterNum {
				lastChapter = rec.ChapterNum
				var inp = rec
				inp.VerseStr = `0`
				inp.VerseNum = 0
				scriptNum++
				inp.ScriptNum = strconv.Itoa(scriptNum)
				key := rec.BookId + `:` + strconv.Itoa(rec.ChapterNum)
				headRec := chapTitle[key]
				inp.UsfmStyle = headRec.UsfmStyle
				inp.ScriptTexts = []string{headRec.ScriptText}
				results = append(results, inp)
			}
			scriptNum++
			rec.ScriptNum = strconv.Itoa(scriptNum)
			rec.ScriptTexts = []string{rec.ScriptText}
			results = append(results, rec)
		}
	}
	return results, status
}
