package mms_asr

import (
	"bufio"
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
	"strings"
)

type MMSASR struct {
	ctx     context.Context
	conn    db.DBAdapter
	lang    string
	sttLang string
	uroman  stdio_exec.StdioExec
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
	lang, status := mms.checkLanguage(a.ctx, a.lang, a.sttLang, "mms_asr")
	if status != nil {
		return status
	}
	pythonScript := filepath.Join(os.Getenv("GOPROJ"), "mms/mms_asr.py")
	writer, reader, status := mms.callStdIOScript(a.ctx, os.Getenv(`FCBH_MMS_ASR_PYTHON`), pythonScript, lang)
	if status != nil {
		return status
	}
	a.uroman, status = stdio_exec.NewStdioExec(a.ctx, os.Getenv(`FCBH_MMS_FA_PYTHON`), uroman.ScriptPath(), "-l", a.lang)
	if status != nil {
		return status
	}
	defer a.uroman.Close()
	for _, file := range files {
		log.Info(a.ctx, "MMS ASR", file.BookId, file.Chapter)
		status = a.processFile(file, writer, reader)
		if status != nil {
			return status
		}
	}
	return status
}

// processFile
func (a *MMSASR) processFile(file input.InputFile, writer *bufio.Writer, reader *bufio.Reader) *log.Status {
	var status *log.Status
	tempDir, err := os.MkdirTemp(os.Getenv(`FCBH_DATASET_TMP`), "mms_asr_")
	if err != nil {
		return log.Error(a.ctx, 500, err, `Error creating temp dir`)
	}
	defer os.RemoveAll(tempDir)
	wavFile, status := ffmpeg.ConvertMp3ToWav(a.ctx, tempDir, file.FilePath())
	if status != nil {
		return status
	}
	var timestamps []db.Audio
	timestamps, status = a.conn.SelectFAScriptTimestamps(file.BookId, file.Chapter)
	if status != nil {
		return status
	}
	timestamps, status = ffmpeg.ChopByTimestamp(a.ctx, tempDir, wavFile, timestamps)
	for i, ts := range timestamps {
		timestamps[i].AudioFile = file.Filename
		timestamps[i].AudioChapterWav = wavFile
		_, err = writer.WriteString(ts.AudioVerseWav + "\n")
		if err != nil {
			return log.Error(a.ctx, 500, err, "Error writing to mms_asr.py")
		}
		err = writer.Flush()
		if err != nil {
			return log.Error(a.ctx, 500, err, "Error flush to mms_asr.py")
		}
		response, err2 := reader.ReadString('\n')
		if err2 != nil {
			return log.Error(a.ctx, 500, err2, `Error reading mms_asr.py response`)
		}
		response = strings.TrimRight(response, "\n")
		fmt.Println(ts.BookId, ts.ChapterNum, ts.VerseStr, ts.ScriptId, response)
		timestamps[i].Text = response
		uRoman, status2 := a.uroman.Process(response)
		if status2 != nil {
			return status2
		}
		timestamps[i].Uroman = uRoman
	}
	//log.Debug(a.ctx, "Finished ASR", file.BookId, file.Chapter)
	var recCount int
	recCount, status = a.conn.UpdateScriptText(timestamps)
	if recCount != len(timestamps) {
		log.Warn(a.ctx, "Timestamp update counts need investigation", recCount, len(timestamps))
	}
	return status
}
