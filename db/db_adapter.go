package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/generic"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/safe"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GetDBPath is not correct with user/project database names
func GetDBPath(database string) string {
	if database == `:memory:` || strings.Contains(database, `/`) {
		return database
	}
	var directory = os.Getenv(`FCBH_DATASET_DB`)
	if directory == `` {
		return database
	} else {
		return filepath.Join(directory, database)
	}
}

// DestroyDatabase should only be used by testing
func DestroyDatabase(database string) {
	var databasePath = GetDBPath(database)
	_, err := os.Stat(databasePath)
	if !os.IsNotExist(err) {
		_ = os.Remove(databasePath)
	}
}

func DatabaseExists(username string, project string) bool {
	database := project + `.db`
	baseDir := os.Getenv(`FCBH_DATASET_DB`)
	if baseDir == `` {
		baseDir = os.Getenv(`HOME`)
	}
	directory := filepath.Join(baseDir, username)
	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(directory, os.ModePerm)
	}
	databasePath := filepath.Join(directory, database)
	_, err = os.Stat(databasePath)
	doesExist := !os.IsNotExist(err)
	return doesExist
}

type DBAdapter struct {
	Ctx          context.Context
	User         string
	Project      string
	Database     string
	DatabasePath string
	DB           *sql.DB
}

// NewerDBAdapter should be used for production
func NewerDBAdapter(ctx context.Context, isNew bool, user string, project string) (DBAdapter, *log.Status) {
	var d DBAdapter
	d.Ctx = ctx
	d.User = user
	d.Project = project
	d.Database = d.Project + ".db"
	baseDir := os.Getenv(`FCBH_DATASET_DB`)
	if baseDir == `` {
		baseDir = os.Getenv(`HOME`)
	}
	directory := filepath.Join(baseDir, d.User)
	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(directory, os.ModePerm)
	}
	d.DatabasePath = filepath.Join(directory, d.Database)
	_, err = os.Stat(d.DatabasePath)
	doesExist := !os.IsNotExist(err)
	if isNew && doesExist {
		_ = os.Remove(d.DatabasePath)
	}
	if !isNew && !doesExist {
		return d, log.Error(ctx, 400, err, `The database does not exist`, d.DatabasePath)
	}
	d.DB, err = sql.Open("sqlite3", d.DatabasePath)
	if err != nil {
		return d, log.Error(ctx, 500, err, `Failed to open database`, d.DatabasePath)
	}
	log.Info(d.Ctx, "DB Opened", d.DatabasePath)
	if isNew {
		createDatabase(d.DB)
	}
	return d, nil
}

// NewDBAdapter should be used for  :memory: database and test.
func NewDBAdapter(ctx context.Context, database string) DBAdapter {
	var databasePath = GetDBPath(database)
	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		log.Fatal(ctx, err)
	}
	var d DBAdapter
	d.Ctx = ctx
	d.Database = database
	d.DatabasePath = databasePath
	d.DB = db
	createDatabase(db)
	return d
}

func createDatabase(db *sql.DB) {
	execDDL(db, `PRAGMA temp_store = MEMORY;`)
	var query = `CREATE TABLE IF NOT EXISTS ident (
		dataset_id INTEGER PRIMARY KEY AUTOINCREMENT,
		bible_id TEXT NOT NULL,
		audio_OT_id TEXT NOT NULL,
		audio_NT_id TEXT NOT NULL,
		text_OT_id TEXT NOT NULL,
		text_NT_id TEXT NOT NULL,
		text_source TEXT NOT NULL,
		language_iso TEXT NOT NULL,
		version_code TEXT NOT NULL,
		languge_id INTEGER NOT NULL,
		rolv_id INTEGER NOT NULL,
		alphabet TEXT NOT NULL,
		language_name TEXT NOT NULL,
		version_name TEXT NOT NULL) STRICT`
	execDDL(db, query)
	query = `CREATE UNIQUE INDEX IF NOT EXISTS ident_bible_idx ON ident (bible_id)`
	execDDL(db, query)
	query = `CREATE TABLE IF NOT EXISTS scripts (
		script_id INTEGER PRIMARY KEY AUTOINCREMENT,
		dataset_id INTEGER NOT NULL,
		book_id TEXT NOT NULL,
		chapter_num INTEGER NOT NULL,
		chapter_end INTEGER NOT NULL,
		verse_str TEXT NOT NULL, /* e.g. 6-10 7,8 6a */
		verse_end TEXT NOT NULL,
		verse_num INTEGER NOT NULL,
		audio_file TEXT NOT NULL, 
		script_num TEXT NOT NULL,
		usfm_style TEXT NOT NULL DEFAULT '',
		person TEXT NOT NULL DEFAULT '',
		actor TEXT NOT NULL DEFAULT '',
		script_text TEXT NOT NULL,
		uroman TEXT NOT NULL DEFAULT '',
		script_begin_ts REAL NOT NULL DEFAULT 0.0,
		script_end_ts REAL NOT NULL DEFAULT 0.0,
		fa_score REAL NOT NULL DEFAULT 0.0,
		FOREIGN KEY(dataset_id) REFERENCES ident(dataset_id)) STRICT`
	execDDL(db, query)
	query = `CREATE UNIQUE INDEX IF NOT EXISTS scripts_idx
		ON scripts (book_id, chapter_num, verse_str)`
	execDDL(db, query)
	query = `CREATE INDEX IF NOT EXISTS scripts_file_idx ON scripts (audio_file)`
	execDDL(db, query)
	query = `CREATE TABLE IF NOT EXISTS words (
		word_id INTEGER PRIMARY KEY AUTOINCREMENT,
		script_id INTEGER NOT NULL,
		word_seq INTEGER NOT NULL,
		verse_num INTEGER NOT NULL,
		ttype TEXT NOT NULL DEFAULT 'W',
		word TEXT NOT NULL,
		uroman TEXT NOT NULL DEFAULT '',
		word_begin_ts REAL NOT NULL DEFAULT 0.0,
		word_end_ts REAL NOT NULL DEFAULT 0.0,
		fa_score REAL NOT NULL DEFAULT 0.0,
		word_enc TEXT NOT NULL DEFAULT '',
		src_word_enc TEXT NOT NULL DEFAULT '', -- planned
		word_multi_enc TEXT NOT NULL DEFAULT '', -- planned
		src_word_multi_enc TEXT NOT NULL DEFAULT '', -- planned
		FOREIGN KEY(script_id) REFERENCES scripts(script_id)) STRICT`
	execDDL(db, query)
	query = `CREATE UNIQUE INDEX IF NOT EXISTS words_idx
		ON words (script_id, word_seq)`
	execDDL(db, query)
	query = `CREATE TABLE IF NOT EXISTS script_mfcc (
		script_id INTEGER PRIMARY KEY,
		rows INTEGER NOT NULL,
		cols INTEGER NOT NULL,
		mfcc_json TEXT NOT NULL,
		FOREIGN KEY (script_id) REFERENCES scripts(script_id)) STRICT`
	execDDL(db, query)
	query = `CREATE TABLE IF NOT EXISTS word_mfcc (
		word_id INTEGER PRIMARY KEY,
		rows INTEGER NOT NULL,
		cols INTEGER NOT NULL,
		mfcc_json TEXT NOT NULL,
		FOREIGN KEY (word_id) REFERENCES words(word_id)) STRICT`
	execDDL(db, query)
	query = `CREATE TABLE IF NOT EXISTS chars (
		char_id INTEGER PRIMARY KEY,
		word_id INTEGER NOT NULL,
		seq INTEGER NOT NULL,
		uroman INTEGER NOT NULL,
		start_ts REAL NOT NULL,
		end_ts REAL NOT NULL,
		fa_score REAL NOT NULL,
		FOREIGN KEY (word_id) REFERENCES words(word_id)) STRICT`
	execDDL(db, query)
}

// CopyDatabase copies a database, closes it and return a connection to the copy
func (d *DBAdapter) CopyDatabase(suffix string) (DBAdapter, *log.Status) {
	var result DBAdapter
	ext := filepath.Ext(d.DatabasePath)
	endName := len(d.DatabasePath) - len(ext)
	targetPath := d.DatabasePath[:endName] + suffix + ext
	d.Close()
	source, err := os.Open(d.DatabasePath)
	if err != nil {
		return result, log.Error(d.Ctx, 500, err, `Error Copying Database step 1`)
	}
	destination, err := os.Create(targetPath)
	if err != nil {
		return result, log.Error(d.Ctx, 500, err, `Error Copying Database step 2`)
	}
	_, err = io.Copy(destination, source)
	if err != nil {
		return result, log.Error(d.Ctx, 500, err, `Error Copying Database step 3`)
	}
	_ = destination.Close()
	_ = source.Close()
	result = *d
	result.DatabasePath = targetPath
	result.Database = filepath.Base(targetPath)
	ext = filepath.Ext(result.Database)
	endName = len(result.Database) - len(ext)
	result.Project = result.Database[:endName]
	result.DB, err = sql.Open("sqlite3", result.DatabasePath)
	if err != nil {
		return result, log.Error(d.Ctx, 500, err, `Error Copying Database step 4`, result.DatabasePath)
	}
	log.Info(d.Ctx, "DB Copied", d.DatabasePath, "to", targetPath)
	return result, nil
}

func (d *DBAdapter) EraseDatabase() {
	execDDL(d.DB, `DELETE FROM ident`)
	execDDL(d.DB, `DELETE FROM scripts`)
	execDDL(d.DB, `DELETE FROM words`)
	execDDL(d.DB, `DELETE FROM script_mfcc`)
	execDDL(d.DB, `DELETE FROM word_mfcc`)
	execDDL(d.DB, `DELETE FROM chars`)
}

func execDDL(db *sql.DB, sql string) {
	_, err := db.Exec(sql)
	if err != nil {
		log.Panic(context.Background(), err, sql)
	}
}

func zeroFill(a string, size int) string {
	num := countDigits(a)
	b := "0000000"[:size-num] + a
	return b
}

func (d *DBAdapter) Close() {
	err := d.DB.Close()
	if err != nil {
		log.Info(d.Ctx, err)
	}
}

func (d *DBAdapter) closeDef(cls io.Closer, desc string) {
	err := cls.Close()
	if err != nil {
		log.Warn(d.Ctx, `Error closing`, desc, err)
	}
}

func (d *DBAdapter) commitDML(tx *sql.Tx, query string) *log.Status {
	err := tx.Commit()
	if err != nil {
		return log.Error(d.Ctx, 500, err, query)
	}
	return nil
}

func countDigits(a string) int {
	for i, ch := range a {
		if ch < '0' || ch > '9' {
			return i
		}
	}
	return len(a)
}

func (d *DBAdapter) CountIdentRows() (int, *log.Status) {
	return d.SelectScalarInt(`SELECT count(*) FROM ident`)
}

func (d *DBAdapter) CountScriptRows() (int, *log.Status) {
	return d.SelectScalarInt(`SELECT count(*) FROM scripts`)
}

func (d *DBAdapter) CountWordRows() (int, *log.Status) {
	return d.SelectScalarInt(`SELECT count(*) FROM words`)
}

func (d *DBAdapter) CountScriptMFCCRows() (int, *log.Status) {
	return d.SelectScalarInt(`SELECT count(*) FROM script_mfcc`)
}

func (d *DBAdapter) CountWordMFCCRows() (int, *log.Status) {
	return d.SelectScalarInt(`SELECT count(*) FROM word_mfcc`)
}

func (d *DBAdapter) DeleteMFCCs() {
	query := `DELETE FROM script_mfcc`
	execDDL(d.DB, query)
	query = `DELETE FROM word_mfcc`
	execDDL(d.DB, query)
}

func (d *DBAdapter) DeleteScripts(bookId string, chapterNum int) *log.Status {
	query := `DELETE FROM scripts WHERE book_id = ? AND chapter_num = ?`
	_, err := d.DB.Exec(query, bookId, chapterNum)
	if err != nil {
		return log.Error(d.Ctx, 500, err, `Error deleting scripts`, bookId, chapterNum)
	}
	return nil
}

func (d *DBAdapter) DeleteWords() {
	execDDL(d.DB, `DELETE FROM words`)
}

func (d *DBAdapter) InsertAudioVerses(bookId string, chapter int, filename string, records []Audio) ([]Audio, *log.Status) {
	//var result []int64
	query := `INSERT INTO scripts(dataset_id, book_id, chapter_num, chapter_end, audio_file,
			script_num, verse_num, verse_str, verse_end, script_text, script_begin_ts, script_end_ts,
			fa_score, uroman) 
			VALUES (1,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	tx, stmt := d.prepareDML(query)
	defer d.closeDef(stmt, "InsertAudioVerse stmt")
	for i, rec := range records {
		scriptNum := zeroFill(rec.VerseStr, 5)
		qry, err := stmt.Exec(bookId, chapter, chapter, filename, scriptNum,
			rec.VerseSeq, rec.VerseStr, rec.VerseEnd, rec.Text, rec.BeginTS, rec.EndTS,
			rec.FAScore, rec.Uroman)
		if err != nil {
			return records, log.Error(d.Ctx, 500, err, `Error while inserting Audio Verse.`)
		}
		records[i].ScriptId, err = qry.LastInsertId()
		if err != nil {
			return records, log.Error(d.Ctx, 500, err, `Error getting lastInsertId, while inserting Audio Verse.`)
		}
	}
	status := d.commitDML(tx, query)
	return records, status
}

func (d *DBAdapter) InsertAudioWords(words []Audio) ([]Audio, *log.Status) {
	query := `REPLACE INTO words(script_id, word_seq, verse_num, word, uroman,
			word_begin_ts, word_end_ts, fa_score)
			VALUES (?,?,?,?,?,?,?,?)`
	tx, stmt := d.prepareDML(query)
	defer d.closeDef(stmt, "InsertAudioWords stmt")
	for i, rec := range words {
		qry, err := stmt.Exec(rec.ScriptId, rec.WordSeq, rec.VerseSeq, rec.Text, rec.Uroman,
			rec.BeginTS, rec.EndTS, rec.FAScore)
		if err != nil {
			return words, log.Error(d.Ctx, 500, err, `Error while inserting Audio Word.`)
		}
		words[i].WordId, err = qry.LastInsertId()
		if err != nil {
			return words, log.Error(d.Ctx, 500, err, `Error getting lastInsertId, while inserting Audio Word.`)
		}
	}
	status := d.commitDML(tx, query)
	return words, status
}

func (d *DBAdapter) InsertAudioChars(words []Audio) *log.Status {
	query := `INSERT INTO chars(word_id, seq, uroman, start_ts, end_ts, fa_score) VALUES (?,?,?,?,?,?)`
	tx, stmt := d.prepareDML(query)
	defer d.closeDef(stmt, `InsertChars stmt`)
	for _, wd := range words {
		for _, ch := range wd.Chars {
			_, err := stmt.Exec(wd.WordId, ch.Seq, ch.Uroman, ch.Start, ch.End, ch.Score)
			if err != nil {
				return log.Error(d.Ctx, 500, err, `Error while inserting Chars.`)
			}
		}
	}
	status := d.commitDML(tx, query)
	return status
}

func (d *DBAdapter) InsertReplaceIdent(id Ident) *log.Status {
	query := `REPLACE INTO ident(dataset_id, bible_id, audio_OT_id, audio_NT_id, text_OT_id, text_NT_id,
		text_source, language_iso, version_code, languge_id, 
		rolv_id, alphabet, language_name, version_name) VALUES (1,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	stmt, err := d.DB.Prepare(query)
	if err != nil {
		return log.Error(d.Ctx, 500, err, `Error while preparing Ident stmt.`)
	}
	defer d.closeDef(stmt, `InsertIdent stmt`)
	_, err = stmt.Exec(id.BibleId, id.AudioOTId, id.AudioNTId, id.TextOTId, id.TextNTId,
		id.TextSource, id.LanguageISO, id.VersionCode, id.LanguageId,
		id.RolvId, id.Alphabet, id.LanguageName, id.VersionName)
	if err != nil {
		return log.Error(d.Ctx, 500, err, `Error while inserting Ident.`)
	}
	return nil
}

func (d *DBAdapter) insertMFCCS(query string, mfccs []MFCC) *log.Status {
	tx, stmt := d.prepareDML(query)
	defer d.closeDef(stmt, "InsertMFCCS stmt")
	for _, rec := range mfccs {
		mfccBytes, err := json.Marshal(rec.MFCC)
		if err != nil {
			return log.Error(d.Ctx, 500, err, `Error converting MFCC to JSON`)
		}
		_, err = stmt.Exec(rec.Id, rec.Rows, rec.Cols, string(mfccBytes))
		if err != nil {
			return log.Error(d.Ctx, 500, err, `Error while inserting MFCC.`)
		}
	}
	status := d.commitDML(tx, query)
	return status
}

func (d *DBAdapter) CheckScriptInserts(records []Script) *log.Status {
	var duplicates []string
	var keyMap = make(map[generic.LineRef]bool)
	for _, rec := range records {
		var key generic.LineRef
		key.BookId = rec.BookId
		key.ChapterNum = rec.ChapterNum
		key.VerseStr = rec.VerseStr
		_, found := keyMap[key]
		if found {
			duplicates = append(duplicates, key.ComposeKey())
		}
		keyMap[key] = true
	}
	if len(duplicates) > 0 {
		return log.ErrorNoErr(d.Ctx, 500, "Duplicate Keys:", strings.Join(duplicates, "\n"))
	}
	return nil
}

func (d *DBAdapter) InsertScripts(records []Script) *log.Status {
	status := d.CheckScriptInserts(records)
	if status != nil {
		return status
	}
	query := `INSERT INTO scripts(dataset_id, book_id, chapter_num, chapter_end, audio_file, script_num, usfm_style, 
			person, actor, verse_num, verse_str, verse_end, script_text, script_begin_ts, script_end_ts) 
			VALUES (1,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	tx, stmt := d.prepareDML(query)
	defer d.closeDef(stmt, "InsertScripts stmt")
	for _, rec := range records {
		rec.ScriptNum = zeroFill(rec.ScriptNum, 5)
		text := safe.SafeStringJoin(rec.ScriptTexts)
		_, err := stmt.Exec(rec.BookId, rec.ChapterNum, rec.ChapterEnd, rec.AudioFile, rec.ScriptNum,
			rec.UsfmStyle, rec.Person, rec.Actor, rec.VerseNum, rec.VerseStr, rec.VerseEnd, text,
			rec.ScriptBeginTS, rec.ScriptEndTS)
		if err != nil {
			return log.Error(d.Ctx, 500, err, `Error while inserting Scripts.`)
		}
	}
	status = d.commitDML(tx, query)
	return status
}

func (d *DBAdapter) InsertScriptMFCCS(mfccs []MFCC) *log.Status {
	query := `REPLACE INTO script_mfcc (script_id, rows, cols, mfcc_json) VALUES (?,?,?,?)`
	return d.insertMFCCS(query, mfccs)
}

func (d *DBAdapter) InsertWordMFCCS(mfccs []MFCC) *log.Status {
	query := `REPLACE INTO word_mfcc (word_id, rows, cols, mfcc_json) VALUES (?,?,?,?)`
	return d.insertMFCCS(query, mfccs)
}

func (d *DBAdapter) InsertWords(records []Word) *log.Status {
	sql1 := `INSERT INTO words(script_id, word_seq, verse_num, ttype, word) VALUES (?,?,?,?,?)`
	tx, stmt := d.prepareDML(sql1)
	defer d.closeDef(stmt, "InsertWords stmt")
	for _, rec := range records {
		_, err := stmt.Exec(rec.ScriptId, rec.WordSeq, rec.VerseNum, rec.TType, rec.Word)
		if err != nil {
			return log.Error(d.Ctx, 500, err, "Error while inserting Words.")
		}
	}
	status := d.commitDML(tx, sql1)
	return status
}

func (d *DBAdapter) prepareDML(query string) (*sql.Tx, *sql.Stmt) {
	tx, err := d.DB.Begin()
	if err != nil {
		log.Fatal(d.Ctx, err, query)
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Fatal(d.Ctx, err, query)
	}
	return tx, stmt
}

func (d *DBAdapter) SelectBookChapter() ([]Script, *log.Status) {
	var results []Script
	query := `SELECT distinct book_id, chapter_num FROM scripts`
	rows, err := d.DB.Query(query)
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, `Error reading rows in SelectBookChapter`)
	}
	defer d.closeDef(rows, `SelectBookChapter`)
	for rows.Next() {
		var scp Script
		err = rows.Scan(&scp.BookId, &scp.ChapterNum)
		if err != nil {
			return results, log.Error(d.Ctx, 500, err, `Error scanning in SelectBookChapter`)
		}
		results = append(results, scp)
	}
	err = rows.Err()
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, `Error at end of rows in SelectBookChapter`)
	}
	return results, nil
}

func (d *DBAdapter) SelectBookChapterFilename() ([]Script, *log.Status) {
	var results []Script
	query := `SELECT distinct book_id, chapter_num, audio_file FROM scripts WHERE audio_file != ''`
	rows, err := d.DB.Query(query)
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, `Error reading rows in SelectBookChapter`)
	}
	defer d.closeDef(rows, `SelectBookChapterFilename`)
	for rows.Next() {
		var scp Script
		err = rows.Scan(&scp.BookId, &scp.ChapterNum, &scp.AudioFile)
		if err != nil {
			return results, log.Error(d.Ctx, 500, err, `Error scanning in SelectBookChapterFilename`)
		}
		results = append(results, scp)
	}
	err = rows.Err()
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, `Error at end of rows in SelectBookChapterFilename`)
	}
	return results, nil
}

func (d *DBAdapter) SelectIdent() (Ident, *log.Status) {
	var results Ident
	query := `SELECT dataset_id, bible_id, audio_OT_id, audio_NT_id, text_OT_id, 
		text_NT_id, text_source, language_iso, version_code, languge_id, 
		rolv_id, alphabet, language_name, version_name 
		FROM ident WHERE dataset_id = 1`
	rows, err := d.DB.Query(query)
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, `Error reading rows in SelectIdent`)
	}
	defer d.closeDef(rows, `SelectIdent`)
	for rows.Next() {
		var id Ident
		err = rows.Scan(&id.DatasetId, &id.BibleId, &id.AudioOTId, &id.AudioNTId, &id.TextOTId,
			&id.TextNTId, &id.TextSource, &id.LanguageISO, &id.VersionCode, &id.LanguageId,
			&id.RolvId, &id.Alphabet, &id.LanguageName, &id.VersionName)
		if err != nil {
			return results, log.Error(d.Ctx, 500, err, `Error scanning in SelectIdent`)
		}
		results = id
	}
	err = rows.Err()
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, `Error at end of rows in SelectIdent`)
	}
	return results, nil
}

// SelectScriptLine selects by script_id and returns one line of script text
func (d *DBAdapter) SelectScriptLine(lineId int64) (string, *log.Status) {
	return d.selectLine(lineId, `SELECT script_text FROM scripts WHERE script_id = ?`)
}

// SelectUromanLine selects by script_id and returns one line of script text
func (d *DBAdapter) SelectUromanLine(lineId int64) (string, *log.Status) {
	return d.selectLine(lineId, `SELECT uroman FROM scripts WHERE script_id = ?`)
}

func (d *DBAdapter) selectLine(lineId int64, query string) (string, *log.Status) {
	var result string
	rows, err := d.DB.Query(query, lineId)
	if err != nil {
		return result, log.Error(d.Ctx, 500, err, `Error reading rows in selectLine`)
	}
	defer d.closeDef(rows, `selectLine`)
	if rows.Next() {
		err = rows.Scan(&result)
		if err != nil {
			return result, log.Error(d.Ctx, 500, err, `Error scanning in selectLine`)
		}
	}
	err = rows.Err()
	if err != nil {
		return result, log.Error(d.Ctx, 500, err, `Error at end of rows in selectLine`)
	}
	return result, nil
}

// SelectScriptsByChapter is used by Compare
func (d *DBAdapter) SelectScriptsByChapter(bookId string, chapterNum int) ([]Script, *log.Status) {
	var results []Script
	sqlStmt := `SELECT script_id, chapter_end, verse_str, verse_end, script_text, uroman, script_begin_ts, script_end_ts FROM scripts 
			WHERE book_id=? AND chapter_num=?
			ORDER BY script_id`
	rows, err := d.DB.Query(sqlStmt, bookId, chapterNum)
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, `Error reading rows in ReadScriptByChapter`)
	}
	defer d.closeDef(rows, `SelectScriptsByChapter`)
	for rows.Next() {
		var vs Script
		vs.BookId = bookId
		vs.ChapterNum = chapterNum
		err = rows.Scan(&vs.ScriptId, &vs.ChapterEnd, &vs.VerseStr, &vs.VerseEnd, &vs.ScriptText, &vs.URoman, &vs.ScriptBeginTS, &vs.ScriptEndTS)
		if err != nil {
			return results, log.Error(d.Ctx, 500, err, `Error scanning in ReadScriptByChapter`)
		}
		results = append(results, vs)
	}
	err = rows.Err()
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, `Error at end of rows in ReadingScriptByChapter`)
	}
	return results, nil
}

func (d *DBAdapter) SelectScalarInt(sql string) (int, *log.Status) {
	var count int
	rows, err := d.DB.Query(sql)
	if err != nil {
		return count, log.Error(d.Ctx, 500, err, sql)
	}
	defer d.closeDef(rows, "SelectScalarInt stmt")
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return count, log.Error(d.Ctx, 500, err, sql)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.Ctx, err, sql)
	}
	return count, nil
}

// SelectScripts is used by WordParser
func (d *DBAdapter) SelectScripts() ([]Script, *log.Status) {
	var results []Script
	query := `SELECT script_id, book_id, chapter_num, script_num, verse_num, verse_str, script_text 
		FROM scripts ORDER BY script_id`
	rows, err := d.DB.Query(query)
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, "Error during select scripts")
	}
	defer d.closeDef(rows, "SelectScripts stmt")
	for rows.Next() {
		var rec Script
		err = rows.Scan(&rec.ScriptId, &rec.BookId, &rec.ChapterNum, &rec.ScriptNum,
			&rec.VerseNum, &rec.VerseStr, &rec.ScriptText)
		if err != nil {
			return results, log.Error(d.Ctx, 500, err, "Error in SelectScripts.")
		}
		rec.ScriptNum = strings.TrimLeft(rec.ScriptNum, "0")
		results = append(results, rec)
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.Ctx, err, query)
	}
	return results, nil
}

// SelectScriptsByBookChapter ...
func (d *DBAdapter) SelectScriptsByBookChapter(bookId string, chapter int) ([]Script, *log.Status) {
	var results []Script
	var query = `SELECT script_id, script_text FROM scripts 
		WHERE book_id = ? AND chapter_num = ? ORDER BY script_id`
	rows, err := d.DB.Query(query, bookId, chapter)
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, "Error during Select Script By Book Chapter.")
	}
	defer d.closeDef(rows, "SelectScriptsByBookChapter stmt")
	for rows.Next() {
		var rec = Script{BookId: bookId, ChapterNum: chapter}
		err = rows.Scan(&rec.ScriptId, &rec.ScriptText)
		if err != nil {
			return results, log.Error(d.Ctx, 500, err, "Error during Select Script By Book Chapter.")
		}
		results = append(results, rec)
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.Ctx, err, query)
	}
	return results, nil
}

func (d *DBAdapter) SelectScriptHeadings() ([]Script, *log.Status) {
	var result []Script
	query := `SELECT script_id, book_id, chapter_num, usfm_style, verse_num, verse_str, script_text 
		FROM scripts
		WHERE usfm_style IN ('para.h', 'para.mt', 'para.mt1', 'para.mt2', 'para.mt3')
		ORDER BY script_id`
	rows, err := d.DB.Query(query)
	if err != nil {
		return result, log.Error(d.Ctx, 500, err, "Error during Select Script Headings.")
	}
	defer d.closeDef(rows, "SelectScriptHeadings stmt")
	for rows.Next() {
		var rec Script
		err = rows.Scan(&rec.ScriptId, &rec.BookId, &rec.ChapterNum, &rec.UsfmStyle, &rec.VerseNum,
			&rec.VerseStr, &rec.ScriptText)
		if err != nil {
			return result, log.Error(d.Ctx, 500, err, "Error during Select Script Headings.")
		}
		result = append(result, rec)
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.Ctx, err, query)
	}
	return result, nil
}

// SelectScriptIds is used by api_dbp_timestamps
func (d *DBAdapter) SelectScriptIds() ([]Script, *log.Status) {
	var results []Script
	query := `SELECT script_id, book_id, chapter_num, script_num, verse_str 
		FROM scripts ORDER BY script_id`
	rows, err := d.DB.Query(query)
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, "Error during select scripts")
	}
	defer d.closeDef(rows, "SelectScriptIds stmt")
	for rows.Next() {
		var rec Script
		err = rows.Scan(&rec.ScriptId, &rec.BookId, &rec.ChapterNum, &rec.ScriptNum, &rec.VerseStr)
		if err != nil {
			return results, log.Error(d.Ctx, 500, err, "Error in SelectScripts.")
		}
		results = append(results, rec)
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.Ctx, err, query)
	}
	return results, nil
}

func (d *DBAdapter) SelectFAScriptTimestamps(bookId string, chapter int) ([]Audio, *log.Status) {
	var results []Audio
	var query = `SELECT script_id, audio_file, verse_str, verse_num, 
			script_text, uroman, script_begin_ts, script_end_ts, fa_score 
			FROM scripts WHERE book_id = ? AND chapter_num = ?
			ORDER BY script_id`
	rows, err := d.DB.Query(query, bookId, chapter)
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, "Error during SelectFAScriptTimestamps By Book Chapter.")
	}
	defer d.closeDef(rows, "SelectFAScriptTimestamps stmt")
	for rows.Next() {
		var rec Audio
		rec.BookId = bookId
		rec.ChapterNum = chapter
		err = rows.Scan(&rec.ScriptId, &rec.AudioFile, &rec.VerseStr, &rec.VerseSeq,
			&rec.Text, &rec.Uroman, &rec.BeginTS, &rec.EndTS, &rec.FAScore)
		if err != nil {
			return results, log.Error(d.Ctx, 500, err, "Error during SelectFAScriptTimestamps By Book Chapter.")
		}
		results = append(results, rec)
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.Ctx, err, query)
	}
	return results, nil
}

func (d *DBAdapter) SelectFACharTimestamps() ([]generic.AlignChar, *log.Status) {
	var chars []generic.AlignChar
	var query = `SELECT s.audio_file, s.script_id, s.book_id, s.chapter_num, s.verse_str,
				w.word_id, w.word, c.char_id, c.seq, c.uroman, c.start_ts, c.end_ts, c.fa_score
				FROM scripts s JOIN words w ON s.script_id = w.script_id
				JOIN chars c ON w.word_id = c.word_id
				WHERE w.ttype = 'W'
				ORDER BY c.char_id`
	rows, err := d.DB.Query(query)
	if err != nil {
		return chars, log.Error(d.Ctx, 500, err, "Error during SelectFACharTimestamps.")
	}
	defer d.closeDef(rows, "SelectFACharTimestamps stmt")
	var ref generic.LineRef
	for rows.Next() {
		var ch generic.AlignChar
		err = rows.Scan(&ch.AudioFile, &ch.LineId, &ref.BookId, &ref.ChapterNum, &ref.VerseStr,
			&ch.WordId, &ch.Word, &ch.CharId, &ch.CharSeq, &ch.Uroman, &ch.BeginTS, &ch.EndTS,
			&ch.FAScore)
		if err != nil {
			return chars, log.Error(d.Ctx, 500, err, "Error in SelectFACharTimestamps.")
		}
		ch.LineRef = ref.ComposeKey()
		chars = append(chars, ch)
	}
	return chars, nil
}

func (d *DBAdapter) SelectFAWordTimestamps() ([]Audio, *log.Status) {
	var results []Audio
	var query = `SELECT w.word_id, w.script_id, s.book_id, s.chapter_num, s.verse_str, 
		s.verse_num, w.word_seq, w.word, w.uroman, w.word_begin_ts, w.word_end_ts, w.fa_score,
		s.script_begin_ts, s.script_end_ts, s.fa_score
		FROM words w JOIN scripts s ON w.script_id = s.script_id
		WHERE w.ttype = 'W'
		ORDER BY w.word_id`
	rows, err := d.DB.Query(query)
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, "Error during SelectFAWordTimestamps By Book Chapter.")
	}
	defer d.closeDef(rows, "SelectFAWordTimestamps stmt")
	for rows.Next() {
		var rec Audio
		err = rows.Scan(&rec.WordId, &rec.ScriptId, &rec.BookId, &rec.ChapterNum, &rec.VerseStr,
			&rec.VerseSeq, &rec.WordSeq, &rec.Text, &rec.Uroman, &rec.BeginTS, &rec.EndTS, &rec.FAScore,
			&rec.ScriptBeginTS, &rec.ScriptEndTS, &rec.ScriptFAScore)
		if err != nil {
			return results, log.Error(d.Ctx, 500, err, "Error during SelectFAWordTimestamps By Book Chapter.")
		}
		results = append(results, rec)
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.Ctx, err, query)
	}
	return results, nil
}

func (d *DBAdapter) SelectScriptTimestamps(bookId string, chapter int) ([]Timestamp, *log.Status) {
	query := `SELECT script_id, verse_str, script_begin_ts, script_end_ts
		FROM scripts WHERE book_id = ? AND chapter_num = ? ORDER BY script_id`
	return d.selectTimestamps(query, bookId, chapter)
}

func (d *DBAdapter) selectTimestamps(query string, bookId string, chapter int) ([]Timestamp, *log.Status) {
	var results []Timestamp
	rows, err := d.DB.Query(query, bookId, chapter)
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, "Error during Select Timestamps By Book Chapter.")
	}
	defer d.closeDef(rows, "SelectTimestamps stmt")
	for rows.Next() {
		var rec Timestamp
		err := rows.Scan(&rec.Id, &rec.VerseStr, &rec.BeginTS, &rec.EndTS)
		if err != nil {
			return results, log.Error(d.Ctx, 500, err, "Error during Select Timestamps By Book Chapter.")
		}
		results = append(results, rec)
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.Ctx, err, query)
	}
	return results, nil
}

func (d *DBAdapter) SelectScriptLineLength() (int, *log.Status) {
	var query = `SELECT IFNULL(CAST(AVG(LENGTH(script_num)) AS INT),0) FROM scripts;`
	return d.SelectScalarInt(query)
}

func (d *DBAdapter) SelectVerseLength() (int, *log.Status) {
	var query = `SELECT IFNULL(CAST(AVG(LENGTH(verse_str)) AS INT),0) FROM scripts;`
	return d.SelectScalarInt(query)
}

// SelectWords is used by encode.FastText
func (d *DBAdapter) SelectWords() ([]Word, *log.Status) {
	var results []Word
	var query = `SELECT word_id, ttype, word FROM words ORDER BY word_id`
	rows, err := d.DB.Query(query)
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, "Error during Select Words.")
	}
	defer d.closeDef(rows, "SelectWords stmt")
	for rows.Next() {
		var rec Word
		err := rows.Scan(&rec.WordId, &rec.TType, &rec.Word)
		if err != nil {
			return results, log.Error(d.Ctx, 500, err, "Error during Select Words.")
		}
		results = append(results, rec)
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.Ctx, err, query)
	}
	return results, nil
}

// SelectWordsByBookChapter is used by Aeneas and mms_fa
func (d *DBAdapter) SelectWordsByBookChapter(bookId string, chapter int) ([]Word, *log.Status) {
	var results []Word
	var query = `SELECT s.verse_str, w.script_id, w.word_id, w.word_seq, w.word
		FROM words w JOIN scripts s ON w.script_id = s.script_id
		WHERE w.ttype = 'W' AND s.book_id = ? AND s.chapter_num = ? ORDER BY w.word_id`
	rows, err := d.DB.Query(query, bookId, chapter)
	if err != nil {
		return results, log.Error(d.Ctx, 500, err, "Error during Select Words By Book Chapter.")
	}
	defer d.closeDef(rows, "SelectWordsByBookChapter stmt")
	for rows.Next() {
		var rec Word
		err = rows.Scan(&rec.VerseStr, &rec.ScriptId, &rec.WordId, &rec.WordSeq, &rec.Word)
		if err != nil {
			return results, log.Error(d.Ctx, 500, err, "Error during Select Words By Book Chapter.")
		}
		results = append(results, rec)
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.Ctx, err, query)
	}
	return results, nil
}

func (d *DBAdapter) SelectWordTimestamps(bookId string, chapter int) ([]Timestamp, *log.Status) {
	query := `SELECT w.word_id, s.verse_str, w.word_begin_ts, w.word_end_ts
		FROM words w JOIN scripts s ON w.script_id = s.script_id
		WHERE w.ttype = 'W' AND s.book_id = ? AND s.chapter_num = ? ORDER BY w.word_id`
	return d.selectTimestamps(query, bookId, chapter)
}

func (d *DBAdapter) UpdateIdent(ident Ident) *log.Status {
	query := `UPDATE ident SET audio_OT_id = ?, audio_NT_id = ?, text_OT_id = ?,
		text_NT_id = ?, text_source = ? WHERE dataset_id = 1`
	stmt, err := d.DB.Prepare(query)
	defer d.closeDef(stmt, `UpdateIdent stmt`)
	if err != nil {
		return log.Error(d.Ctx, 500, err, `Error while preparing Ident stmt.`)
	}
	_, err = stmt.Exec(ident.AudioOTId, ident.AudioNTId, ident.TextOTId, ident.TextNTId,
		ident.TextSource)
	if err != nil {
		return log.Error(d.Ctx, 500, err, `Error while updating Ident.`)
	}
	return nil
}

func (d *DBAdapter) UpdateScriptTimestamps(scripts []Timestamp) *log.Status {
	query := `UPDATE scripts SET audio_file = ?, script_begin_ts = ?,
		script_end_ts = ? WHERE script_id = ?`
	tx, stmt := d.prepareDML(query)
	defer d.closeDef(stmt, "UpdateScriptTimestamps stmt")
	for _, rec := range scripts {
		_, err := stmt.Exec(rec.AudioFile, rec.BeginTS, rec.EndTS, rec.Id)
		if err != nil {
			return log.Error(d.Ctx, 500, err, `Error while updating script timestamps.`)
		}
	}
	status := d.commitDML(tx, query)
	return status
}

func (d *DBAdapter) UpdateEraseScriptText() *log.Status {
	query := `UPDATE scripts SET script_text = "", uroman = ""`
	tx, stmt := d.prepareDML(query)
	defer d.closeDef(stmt, "UpdateEraseScriptText stmt")
	_, err := stmt.Exec()
	if err != nil {
		return log.Error(d.Ctx, 500, err, `Error while updating script text.`)
	}
	status := d.commitDML(tx, query)
	if status != nil {
		return status
	}
	execDDL(d.DB, `DELETE FROM words`)
	execDDL(d.DB, `DELETE FROM word_mfcc`)
	execDDL(d.DB, `DELETE FROM chars`)
	return nil
}

func (d *DBAdapter) UpdateUromanText(scripts []Script) (int, *log.Status) {
	var rowsUpdated int64
	query := `UPDATE scripts SET uroman = ? WHERE script_id = ?`
	tx, stmt := d.prepareDML(query)
	defer d.closeDef(stmt, "UpdateUromanText stmt")
	for _, rec := range scripts {
		res, err := stmt.Exec(rec.URoman, rec.ScriptId)
		if err != nil {
			return int(rowsUpdated), log.Error(d.Ctx, 500, err, `Error while updating uroman text.`)
		}
		affected, _ := res.RowsAffected()
		rowsUpdated += affected
	}
	status := d.commitDML(tx, query)
	return int(rowsUpdated), status
}

func (d *DBAdapter) UpdateScriptText(audio []Audio) (int, *log.Status) {
	var rowsUpdated int64
	query := `UPDATE scripts SET script_text = ?, uroman = ? WHERE script_id = ?`
	tx, stmt := d.prepareDML(query)
	defer d.closeDef(stmt, "UpdateScriptText stmt")
	for _, rec := range audio {
		res, err := stmt.Exec(rec.Text, rec.Uroman, rec.ScriptId)
		if err != nil {
			return int(rowsUpdated), log.Error(d.Ctx, 500, err, `Error while updating script text.`)
		}
		affected, _ := res.RowsAffected()
		rowsUpdated += affected
	}
	status := d.commitDML(tx, query)
	return int(rowsUpdated), status
}

func (d *DBAdapter) UpdateScriptFATimestamps(audio []Audio) *log.Status {
	query := `UPDATE scripts SET audio_file = ?, script_begin_ts = ?,
		script_end_ts = ?, fa_score = ?, uroman = ? WHERE script_id = ?`
	tx, stmt := d.prepareDML(query)
	defer d.closeDef(stmt, "UpdateScriptFATimestamps stmt")
	var rowsUpdated int64
	for _, rec := range audio {
		res, err := stmt.Exec(rec.AudioFile, rec.BeginTS, rec.EndTS, rec.FAScore, rec.Uroman, rec.ScriptId)
		if err != nil {
			return log.Error(d.Ctx, 500, err, `Error while updating script FA timestamps.`)
		}
		affected, _ := res.RowsAffected()
		rowsUpdated += affected
	}
	status := d.commitDML(tx, query)
	if status != nil {
		return status
	}
	if int(rowsUpdated) != len(audio) {
		return log.ErrorNoErr(d.Ctx, 400, strconv.Itoa(len(audio))+" rows updated "+strconv.Itoa(int(rowsUpdated)))
	}
	return nil
}

func (d *DBAdapter) UpdateWordFATimestamps(audio []Audio) *log.Status {
	query := `UPDATE words SET word_begin_ts = ?,
		word_end_ts = ?, fa_score = ?, uroman = ? WHERE word_id = ?`
	tx, stmt := d.prepareDML(query)
	defer d.closeDef(stmt, "UpdateWordFATimestamps stmt")
	var rowsUpdated int64
	for _, rec := range audio {
		res, err := stmt.Exec(rec.BeginTS, rec.EndTS, rec.FAScore, rec.Uroman, rec.WordId)
		if err != nil {
			return log.Error(d.Ctx, 500, err, `Error while updating word FA timestamps.`)
		}
		affected, _ := res.RowsAffected()
		rowsUpdated += affected
	}
	status := d.commitDML(tx, query)
	if status != nil {
		return status
	}
	if int(rowsUpdated) != len(audio) {
		return log.ErrorNoErr(d.Ctx, 400, strconv.Itoa(len(audio))+" rows updated "+strconv.Itoa(int(rowsUpdated)))
	}
	return nil
}

func (d *DBAdapter) UpdateWordEncodings(words []Word) *log.Status {
	query := `UPDATE words SET word_enc = ? WHERE word_id = ?`
	tx, stmt := d.prepareDML(query)
	defer d.closeDef(stmt, "UpdateWordEncodings stmt")
	for _, rec := range words {
		encBytes, err := json.Marshal(rec.WordEncoded)
		if err != nil {
			return log.Error(d.Ctx, 500, err, `Error converting word enc to JSON`)
		}
		encStr := string(encBytes)
		if encStr != `null` {
			_, err = stmt.Exec(encStr, rec.WordId)
			if err != nil {
				return log.Error(d.Ctx, 500, err, `Error while inserting word enc.`)
			}
		}
	}
	status := d.commitDML(tx, query)
	return status
}

func (d *DBAdapter) UpdateWordTimestamps(words []Timestamp) *log.Status {
	query := `UPDATE words SET word_begin_ts = ?, word_end_ts = ? WHERE word_id = ?`
	tx, stmt := d.prepareDML(query)
	defer d.closeDef(stmt, "UpdateWordTimestamps stmt")
	for _, rec := range words {
		_, err := stmt.Exec(rec.BeginTS, rec.EndTS, rec.Id)
		if err != nil {
			return log.Error(d.Ctx, 500, err, `Error while updating word timestamps.`)
		}
	}
	status := d.commitDML(tx, query)
	return status
}
