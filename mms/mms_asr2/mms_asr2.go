package mms_asr

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/mms"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/ffmpeg"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/stdio_exec"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/uroman"
	"os"
	"path/filepath"
	"strconv"
)

type MMSASR2 struct {
	ctx      context.Context
	conn     db.DBAdapter
	lang     string
	sttLang  string
	adapter  bool
	mmsAsrPy *stdio_exec.StdioExec
	uroman   *stdio_exec.StdioExec
}

func NewMMSASR2(ctx context.Context, conn db.DBAdapter, lang string, sttLang string, adapter bool) MMSASR2 {
	var a MMSASR2
	a.ctx = ctx
	a.conn = conn
	a.lang = lang
	a.sttLang = sttLang
	a.adapter = adapter
	return a
}

// ProcessFiles will perform Auto Speech Recognition on these files
func (a *MMSASR2) ProcessFiles(files []input.InputFile) (status *log.Status) {
	tempDir, err := os.MkdirTemp(os.Getenv(`FCBH_DATASET_TMP`), "mms_asr_")
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
	defer func() {
		status = a.mmsAsrPy.Close()
	}()
	a.uroman, status = stdio_exec.NewStdioExec(a.ctx, os.Getenv(`FCBH_MMS_FA_PYTHON`), uroman.ScriptPath(), "-l", a.lang)
	if status != nil {
		return status
	}
	defer func() {
		status = a.uroman.Close()
	}()
	for _, file := range files {
		status = a.processFile(file, tempDir)
		if status != nil {
			return status
		}
	}
	return status
}

// processFile
func (a *MMSASR2) processFile(file input.InputFile, tempDir string) *log.Status {
	var status *log.Status
	fmt.Println("Process", file.FilePath())
	wavFile, status := ffmpeg.ConvertMp3ToWav(a.ctx, tempDir, file.FilePath())
	if status != nil {
		return status
	}
	var audioFile db.Audio
	audioFile.AudioFile = file.Filename
	audioFile.AudioChapterWav = wavFile
	response, status1 := a.mmsAsrPy.Process(wavFile)
	if status1 != nil {
		return status1
	}
	fmt.Println("response:", response)
	err := os.WriteFile(file.MediaId+"_"+file.BookId+strconv.Itoa(file.Chapter), []byte(response), 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	audioFile.Text = response
	uRoman, status2 := a.uroman.Process(response)
	if status2 != nil {
		return status2
	}
	audioFile.Uroman = uRoman
	var recCount int
	recCount, status = a.conn.UpdateScriptText([]db.Audio{audioFile})
	if recCount != 1 {
		log.Warn(a.ctx, "ASR update counts needs investigation", recCount, 1)
	}
	return status
}

func (a *MMSASR2) selectScriptLine(scriptLine string) (db.Audio, *log.Status) {
	var rec db.Audio
	var query = `SELECT script_id, book_id, chapter_num, audio_file, script_begin_ts, script_end_ts FROM scripts WHERE script_num = ?`
	row := a.conn.DB.QueryRow(query, scriptLine)
	err := row.Scan(&rec.ScriptId, &rec.BookId, &rec.ChapterNum, &rec.AudioFile, &rec.ScriptBeginTS, &rec.ScriptEndTS)
	if err != nil {
		return rec, log.Error(a.ctx, 500, err, "Error during SelectScriptLine.")
	}
	return rec, nil
}
