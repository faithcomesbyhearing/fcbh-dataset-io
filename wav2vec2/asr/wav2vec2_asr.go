package asr

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/ffmpeg"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/stdio_exec"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/uroman"
	"os"
	"path/filepath"
)

type Wav2Vec2ASR struct {
	ctx     context.Context
	conn    db.DBAdapter
	lang    string
	sttLang string
	adapter bool
	asrPy   stdio_exec.StdioExec
	uroman  stdio_exec.StdioExec
}

func NewWav2Vec2ASR(ctx context.Context, conn db.DBAdapter, lang string, sttLang string) Wav2Vec2ASR {
	var a Wav2Vec2ASR
	a.ctx = ctx
	a.conn = conn
	a.lang = lang
	a.sttLang = sttLang
	return a
}

// ProcessFiles will perform Auto Speech Recognition on these files
func (a *Wav2Vec2ASR) ProcessFiles(files []input.InputFile) *log.Status {
	var status *log.Status
	tempDir, err := os.MkdirTemp(os.Getenv(`FCBH_DATASET_TMP`), "wav2vec2_asr_")
	if err != nil {
		return log.Error(a.ctx, 500, err, `Error creating temp dir`)
	}
	defer os.RemoveAll(tempDir)
	var lang = a.lang
	if a.sttLang != "" {
		lang = a.sttLang
	}
	//if !a.adapter {
	//	lang, status = mms.CheckLanguage(a.ctx, a.lang, a.sttLang, "mms_asr")
	//	if status != nil {
	//		return status
	//	}
	//}
	status = a.conn.UpdateASRLanguage(lang)
	if status != nil {
		return status
	}
	pythonScript := filepath.Join(os.Getenv("GOPROJ"), "wav2vec2/asr/wav2vec2_asr.py")
	//var useAdapter string
	//if a.adapter {
	//	useAdapter = "adapter"
	//}
	a.asrPy, status = stdio_exec.NewStdioExec(a.ctx, os.Getenv(`FCBH_MMS_ASR_PYTHON`), pythonScript, lang)
	if status != nil {
		return status
	}
	defer a.asrPy.Close()
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
func (a *Wav2Vec2ASR) processFile(file input.InputFile, tempDir string) *log.Status {
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
		response, status1 := a.asrPy.Process(ts.AudioVerseWav)
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

func (a *Wav2Vec2ASR) selectScriptLine(scriptLine string) (db.Audio, *log.Status) {
	var rec db.Audio
	var query = `SELECT script_id, book_id, chapter_num, audio_file FROM scripts WHERE script_num = ?`
	row := a.conn.DB.QueryRow(query, scriptLine)
	err := row.Scan(&rec.ScriptId, &rec.BookId, &rec.ChapterNum, &rec.AudioFile)
	if err != nil {
		return rec, log.Error(a.ctx, 500, err, "Error during SelectScriptLine.")
	}
	return rec, nil
}
