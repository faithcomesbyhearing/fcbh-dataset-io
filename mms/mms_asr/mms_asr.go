package mms_asr

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/mms"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/ffmpeg"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/stdio_exec"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/uroman"
)

type MMSASR struct {
	ctx      context.Context
	conn     db.DBAdapter
	lang     string
	sttLang  string
	adapter  bool
	mmsAsrPy *stdio_exec.StdioExec
	uroman   *stdio_exec.StdioExec
}

func NewMMSASR(ctx context.Context, conn db.DBAdapter, lang string, sttLang string, adapter bool) MMSASR {
	var a MMSASR
	a.ctx = ctx
	a.conn = conn
	a.lang = lang
	a.sttLang = sttLang
	a.adapter = adapter
	return a
}

// ProcessFiles will perform Auto Speech Recognition on these files
func (a *MMSASR) ProcessFiles(files []input.InputFile) *log.Status {
	var status *log.Status
	tempDir, err := os.MkdirTemp(os.Getenv(`FCBH_DATASET_TMP`), "mms_asr_")
	if err != nil {
		return log.Error(a.ctx, 500, err, `Error creating temp dir`)
	}
	//defer os.RemoveAll(tempDir)
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
	pythonScript := filepath.Join(os.Getenv("GOPROJ"), "mms/mms_asr/mms_asr.py")
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
	for _, file := range files {
		status = a.processFile(file, tempDir)
		if status != nil {
			return status
		}
	}
	return status
}

type MMSASR_Input struct {
	Path    string  `json:"path"`
	BeginTS float64 `json:"begin_ts"`
	EndTS   float64 `json:"end_ts"`
}

// processFile
func (a *MMSASR) processFile(file input.InputFile, tempDir string) *log.Status {
	var status *log.Status
	wavFile, status := ffmpeg.ConvertMp3ToWav(a.ctx, tempDir, file.FilePath())
	if status != nil {
		return status
	}
	var audioFiles []db.Audio
	if file.ScriptLine != "" {
		var audioFile db.Audio
		audioFile, status = a.selectScriptLine(file.ScriptLine)
		if status != nil {
			return status
		}
		if audioFile.ScriptEndTS == 0.0 {
			return nil
		}
		log.Info(a.ctx, "MMS ASR", audioFile.BookId, audioFile.ChapterNum, file.ScriptLine)
		audioFile.AudioVerseWav = wavFile
		audioFiles = append(audioFiles, audioFile)
	} else {
		log.Info(a.ctx, "MMS ASR", file.BookId, file.Chapter)
		audioFiles, status = a.conn.SelectFAScriptTimestamps(file.BookId, file.Chapter)
		if status != nil {
			return status
		}
	}
	for i, ts := range audioFiles {
		audioFiles[i].AudioFile = file.Filename
		audioFiles[i].AudioChapterWav = wavFile
		jsonInput := MMSASR_Input{Path: wavFile, BeginTS: ts.BeginTS, EndTS: ts.EndTS}
		jsonBytes, err := json.Marshal(jsonInput)
		if err != nil {
			return log.Error(a.ctx, 500, err, "Error marshalling input to JSON.")
		}
		response, status1 := a.mmsAsrPy.Process(string(jsonBytes))
		if status1 != nil {
			return status1
		}
		log.Info(a.ctx, "Response:", response)
		audioFiles[i].Text = response
		uRoman, status2 := a.uroman.Process(response)
		if status2 != nil {
			return status2
		}
		audioFiles[i].Uroman = uRoman
	}
	var recCount int
	recCount, status = a.conn.UpdateScriptText(audioFiles)
	if recCount != len(audioFiles) {
		log.Warn(a.ctx, "Timestamp update counts needs investigation", recCount, len(audioFiles))
	}
	return status
}

func (a *MMSASR) selectScriptLine(scriptLine string) (db.Audio, *log.Status) {
	var rec db.Audio
	var query = `SELECT script_id, book_id, chapter_num, audio_file, script_begin_ts, script_end_ts FROM scripts WHERE script_num = ?`
	row := a.conn.DB.QueryRow(query, scriptLine)
	err := row.Scan(&rec.ScriptId, &rec.BookId, &rec.ChapterNum, &rec.AudioFile, &rec.ScriptBeginTS, &rec.ScriptEndTS)
	if err != nil {
		return rec, log.Error(a.ctx, 500, err, "Error during SelectScriptLine.")
	}
	return rec, nil
}
