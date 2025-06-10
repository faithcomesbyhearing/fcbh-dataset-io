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
)

type MMSASR struct {
	ctx      context.Context
	conn     db.DBAdapter
	lang     string
	sttLang  string
	mmsAsrPy stdio_exec.StdioExec
	uroman   stdio_exec.StdioExec
}

func NewMMSASR(ctx context.Context, conn db.DBAdapter, lang string, sttLang string) MMSASR {
	var a MMSASR
	a.ctx = ctx
	a.conn = conn
	a.lang = lang
	a.sttLang = sttLang
	return a
}

// ProcessFiles will perform Auto Speech Recognition on these files
func (a *MMSASR) ProcessFiles(files []input.InputFile) *log.Status {
	tempDir, err := os.MkdirTemp(os.Getenv(`FCBH_DATASET_TMP`), "mms_asr_")
	if err != nil {
		return log.Error(a.ctx, 500, err, `Error creating temp dir`)
	}
	defer os.RemoveAll(tempDir)
	lang, status := mms.CheckLanguage(a.ctx, a.lang, a.sttLang, "mms_asr")
	if status != nil {
		return status
	}
	status = a.conn.UpdateASRLanguage(lang)
	if status != nil {
		return status
	}
	pythonScript := filepath.Join(os.Getenv("GOPROJ"), "mms/mms_asr/mms_asr.py")
	a.mmsAsrPy, status = stdio_exec.NewStdioExec(a.ctx, os.Getenv(`FCBH_MMS_ASR_PYTHON`), pythonScript, lang)
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
		log.Info(a.ctx, "MMS ASR", file.BookId, file.Chapter)
		status = a.processFile(file, tempDir)
		if status != nil {
			return status
		}
	}
	return status
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
		audioFile.AudioVerseWav = wavFile
		audioFiles = append(audioFiles, audioFile)
	} else {
		audioFiles, status = a.conn.SelectFAScriptTimestamps(file.BookId, file.Chapter)
		if status != nil {
			return status
		}
		audioFiles, status = ffmpeg.ChopByTimestamp(a.ctx, tempDir, wavFile, audioFiles)
	}
	for i, ts := range audioFiles {
		audioFiles[i].AudioFile = file.Filename
		audioFiles[i].AudioChapterWav = wavFile
		response, status1 := a.mmsAsrPy.Process(ts.AudioVerseWav)
		if status1 != nil {
			return status1
		}
		fmt.Println(ts.BookId, ts.ChapterNum, ts.VerseStr, ts.ScriptId, response)
		audioFiles[i].Text = response
		uRoman, status2 := a.uroman.Process(response)
		if status2 != nil {
			return status2
		}
		audioFiles[i].Uroman = uRoman
	}
	//log.Debug(a.ctx, "Finished ASR", file.BookId, file.Chapter)
	var recCount int
	recCount, status = a.conn.UpdateScriptText(audioFiles)
	if recCount != len(audioFiles) {
		log.Warn(a.ctx, "Timestamp update counts need investigation", recCount, len(audioFiles))
	}
	return status
}

func (a *MMSASR) selectScriptLine(scriptLine string) (db.Audio, *log.Status) {
	var rec db.Audio
	var query = `SELECT script_id, book_id, chapter_num, audio_file FROM scripts WHERE script_num = ?`
	row := a.conn.DB.QueryRow(query, scriptLine)
	err := row.Scan(&rec.ScriptId, &rec.BookId, &rec.ChapterNum, &rec.AudioFile)
	if err != nil {
		return rec, log.Error(a.ctx, 500, err, "Error during SelectScriptLine.")
	}
	return rec, nil
}
