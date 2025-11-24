package mms_asr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/mms"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/ffmpeg"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/stdio_exec"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/uroman"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type MMSASR2 struct {
	ctx          context.Context
	conn         db.DBAdapter
	lang         string
	sttLang      string
	adapter      bool
	mmsAsrPy     *stdio_exec.StdioExec
	uroman       *stdio_exec.StdioExec
	diffMatch    *diffmatchpatch.DiffMatchPatch
	versePattern *regexp.Regexp
}

func NewMMSASR2(ctx context.Context, conn db.DBAdapter, lang string, sttLang string, adapter bool) MMSASR2 {
	var a MMSASR2
	a.ctx = ctx
	a.conn = conn
	a.lang = lang
	a.sttLang = sttLang
	a.adapter = adapter
	a.diffMatch = diffmatchpatch.New()
	a.versePattern = regexp.MustCompile(`\{(\d+)\}`)
	return a
}

// ProcessFiles will perform Auto Speech Recognition on these files
func (a *MMSASR2) ProcessFiles(files []input.InputFile) *log.Status {
	var status *log.Status
	tempDir, err := os.MkdirTemp(os.Getenv(`FCBH_DATASET_TMP`), "mms_asr2_")
	if err != nil {
		return log.Error(a.ctx, 500, err, `Error creating temp dir`)
	}
	defer os.RemoveAll(tempDir)
	var lang = a.lang
	if a.sttLang != "" {
		lang = a.sttLang
	}
	if !a.adapter {
		lang, status = mms.CheckLanguage(a.ctx, a.lang, a.sttLang, "mms_asr")
		if status != nil {
			return status
		}
	}
	status = a.conn.UpdateASRLanguage(lang)
	if status != nil {
		return status
	}
	pythonScript := filepath.Join(os.Getenv("GOPROJ"), "mms/mms_asr2/mms_asr2.py")
	var useAdapter string
	if a.adapter {
		useAdapter = "adapter"
	}
	a.mmsAsrPy, status = stdio_exec.NewStdioExec(a.ctx, os.Getenv(`FCBH_MMS_ASR_PYTHON`), pythonScript, lang, useAdapter)
	if status != nil {
		return status
	}
	defer a.mmsAsrPy.Close()
	a.uroman, status = stdio_exec.NewStdioExec(a.ctx, os.Getenv(`FCBH_MMS_FA_PYTHON`), uroman.ScriptPath(), "-l", a.lang)
	if status != nil {
		return status
	}
	defer a.uroman.Close()
	var response string
	for _, file := range files {
		response, status = a.processASR(file, tempDir)
		if status != nil {
			return status
		}
		fmt.Println("response:", response)
		status = a.parseResult(file, response)
		if status != nil {
			return status
		}
	}
	return status
}

// processFile
func (a *MMSASR2) processASR(file input.InputFile, tempDir string) (string, *log.Status) {
	var response string
	var status *log.Status
	fmt.Println("Process", file.FilePath())
	wavFile, status := ffmpeg.ConvertMp3ToWav(a.ctx, tempDir, file.FilePath())
	if status != nil {
		return response, status
	}
	var audioFile db.Audio
	audioFile.AudioFile = file.Filename
	audioFile.AudioChapterWav = wavFile
	response, status1 := a.mmsAsrPy.Process(wavFile)
	if status1 != nil {
		return response, status1
	}
	// Temp code for debugging
	err := os.WriteFile(file.MediaId+"_"+file.BookId+strconv.Itoa(file.Chapter), []byte(response), 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return response, nil
}

type asrScript struct {
	scriptId int64
	verseStr string
	text     string
	uRoman   string
}

func (a *MMSASR2) parseResult(file input.InputFile, response string) *log.Status {
	scripts, status := a.selectVersesByBookChapter(file.BookId, file.Chapter)
	if status != nil {
		return status
	}
	sourceText := a.combineVerses(scripts)
	respVerses := a.parseASRByOriginal(sourceText, response)
	for i := range respVerses {
		respVerses[i].uRoman, status = a.uroman.Process(respVerses[i].text)
		if status != nil {
			return status
		}
	}
	status = a.ensureASRTable()
	if status != nil {
		return status
	}
	status = a.insertASRText(respVerses)
	if status != nil {
		return status
	}
	return nil
}

func (a *MMSASR2) selectVersesByBookChapter(bookId string, chapter int) ([]asrScript, *log.Status) {
	var results []asrScript
	var query = `SELECT s.script_id, s.verse_str, LOWER(GROUP_CONCAT(w.word, ' ')) AS text
	FROM scripts s JOIN words w ON w.script_id = s.script_id
	WHERE w.ttype = 'W' AND s.book_id = ? AND s.chapter_num = ?
	GROUP BY s.script_id, s.verse_str
	ORDER BY s.script_id, s.verse_str`
	rows, err := a.conn.DB.Query(query, bookId, chapter)
	if err != nil {
		return results, log.Error(a.ctx, 500, err, query, bookId, chapter)
	}
	defer rows.Close()
	for rows.Next() {
		var s asrScript
		err = rows.Scan(&s.scriptId, &s.verseStr, &s.text)
		if err != nil {
			return results, log.Error(a.ctx, 500, err, query, bookId, chapter)
		}
		results = append(results, s)
	}
	err = rows.Err()
	if err != nil {
		return results, log.Error(a.ctx, 500, err, query, bookId, chapter)
	}
	return results, nil
}

func (a *MMSASR2) combineVerses(scripts []asrScript) string {
	var results []string
	for _, s := range scripts {
		results = append(results, "{"+strconv.FormatInt(s.scriptId, 10)+"}")
		results = append(results, s.text)
	}
	return strings.Join(results, "")
}

func (a *MMSASR2) parseASRByOriginal(sourceText string, response string) []asrScript {
	var results []asrScript
	diffs := a.diffMatch.DiffMain(sourceText, response, false)
	var currId string
	var currText []string
	for _, d := range diffs {
		if d.Type == diffmatchpatch.DiffDelete {
			matches := a.versePattern.FindStringSubmatch(d.Text)
			if len(matches) > 0 {
				if len(currText) > 0 {
					var script asrScript
					script.scriptId, _ = strconv.ParseInt(currId, 10, 64)
					script.text = strings.Join(currText, "")
					results = append(results, script)
					currText = nil
				}
				currId = matches[1]
			}
		} else {
			currText = append(currText, d.Text)
		}
	}
	if len(currText) > 0 {
		var script asrScript
		script.scriptId, _ = strconv.ParseInt(currId, 10, 64)
		script.text = strings.Join(currText, "")
		results = append(results, script)
	}
	return results
}

func (a *MMSASR2) ensureASRTable() *log.Status {
	query := `CREATE TABLE IF NOT EXISTS asr (
		script_id INTEGER PRIMARY KEY,
		script_text TEXT NOT NULL,
		uroman TEXT NOT NULL DEFAULT '')`
	_, err := a.conn.DB.Exec(query)
	if err != nil {
		return log.Error(a.ctx, 500, err, query)
	}
	return nil
}

func (a *MMSASR2) insertASRText(scripts []asrScript) *log.Status {
	_, err := a.conn.DB.Exec(`DELETE FROM asr`)
	if err != nil {
		return log.Error(a.ctx, 500, err, "could not delete asr")
	}
	query := `INSERT INTO asr (script_id, script_text, uroman) VALUES (?,?,?)`
	tx, err := a.conn.DB.Begin()
	if err != nil {
		return log.Error(a.ctx, 500, err, query)
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		return log.Error(a.ctx, 500, err, query)
	}
	defer stmt.Close()
	for _, rec := range scripts {
		_, err = stmt.Exec(rec.scriptId, rec.text, rec.uRoman)
		if err != nil {
			return log.Error(a.ctx, 500, err, `Error while inserting asr text.`)
		}
	}
	err = tx.Commit()
	if err != nil {
		return log.Error(a.ctx, 500, err, query)
	}
	return nil
}
