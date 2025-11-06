package mms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/ffmpeg"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Documentation for this implementation of forced alignment
// https://github.com/facebookresearch/fairseq/tree/main/examples/mms/data_prep

// forced_align: time python align_and_segment.py --audio $HOME/MRK.1.wav --text_filepath $HOME/MRK.1.txt --lang eng --outdir $HOME/Desktop/top --uroman $HOME

type ForcedAlign struct {
	ctx      context.Context
	conn     db.DBAdapter // This database adapter must contain the text to be processed
	lang     string
	sttLang  string // I don't know if this is useful
	replacer *strings.Replacer
}

func NewForcedAlign(ctx context.Context, conn db.DBAdapter, lang string, sttLang string) ForcedAlign {
	var f ForcedAlign
	f.ctx = ctx
	f.conn = conn
	f.lang = lang
	f.sttLang = sttLang
	f.replacer = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ")
	return f
}

// ProcessFiles will perform Forced Alignment on these files
func (f *ForcedAlign) ProcessFiles(files []input.InputFile) *log.Status {
	//lang, status := CheckLanguage(f.ctx, f.lang, f.sttLang, "mms_asr") // is this correct for mms_fa
	//if status != nil {
	//	return status
	//}
	var status *log.Status
	lang := f.lang
	if f.sttLang != "" {
		lang = f.sttLang
	}
	for _, file := range files {
		log.Info(f.ctx, "Word FA", file.BookId, file.Chapter)
		status = f.processFile(file, lang)
		if status != nil {
			return status
		}
	}
	return status
}

// processFile will process one audio file through mms forced alignment
func (f *ForcedAlign) processFile(file input.InputFile, lang string) *log.Status {
	var status *log.Status
	tempDir, err := os.MkdirTemp(os.Getenv(`FCBH_DATASET_TMP`), "mms_forced_align_")
	if err != nil {
		return log.Error(f.ctx, 500, err, `Error creating temp dir`)
	}
	defer os.RemoveAll(tempDir)
	wavAudioFile, status := ffmpeg.ConvertMp3ToWav(f.ctx, tempDir, file.FilePath())
	if status != nil {
		return status
	}
	var verses []db.Script
	verses, status = f.conn.SelectScriptsByChapter(file.BookId, file.Chapter)
	if status != nil {
		return status
	}
	var verseRef []int64
	var verseText []string
	for _, vers := range verses {
		verseRef = append(verseRef, int64(vers.ScriptId))
		verseText = append(verseText, f.replacer.Replace(vers.ScriptText))
	}
	textFilePath := filepath.Join(tempDir, "textInput.txt")
	err = os.WriteFile(textFilePath, []byte(strings.Join(verseText, "\n")), 0644)
	if err != nil {
		return log.Error(f.ctx, 500, err, `Error creating text file`)
	}
	outputFile, status := f.forcedAlign(wavAudioFile, textFilePath, lang, tempDir)
	if status != nil {
		return status
	}
	status = f.processPyOutput(file, outputFile, verseRef)
	return status
}

func (f *ForcedAlign) forcedAlign(audioFile string, textFile string, lang string, tempDir string) (string, *log.Status) {
	var result string
	var status *log.Status
	MMSFAPYTHON := os.Getenv("FCBH_MMS_FA_PYTHON")
	pythonScript := filepath.Join(os.Getenv("GOPROJ"), "/mms/forced_align/align_and_segment.py")
	outputDir := filepath.Join(tempDir, `output`)
	cmd := exec.Command(MMSFAPYTHON,
		pythonScript,
		`--audio`, audioFile,
		`--text_filepath`, textFile,
		`--lang`, lang,
		`--outdir`, outputDir,
		`--uroman`, filepath.Dir(MMSFAPYTHON))
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	if err != nil {
		status = log.Error(f.ctx, 500, err, stderrBuf.String())
	}
	result = filepath.Join(outputDir, "manifest.json")
	return result, status
}

type FAOutput struct {
	AudioStart     float64 `json:"audio_start_sec"`
	AudioFilePath  string  `json:"audio_filepath"`
	Duration       float64 `json:"duration"`
	Text           string  `json:"text"`
	NormalizedText string  `json:"normalized_text"`
	UromanTokens   string  `json:"uroman_tokens"`
}

func (f *ForcedAlign) processPyOutput(file input.InputFile, outputFile string, references []int64) *log.Status {
	var status *log.Status
	content, err := os.ReadFile(outputFile)
	if err != nil {
		return log.Error(f.ctx, 500, err, `Error reading output file`)
	}
	content = bytes.Trim(content, "\n")
	lines := bytes.Split(content, []byte("\n"))
	if len(lines) != len(references) {
		return log.ErrorNoErr(f.ctx, 400, "output len="+strconv.Itoa(len(lines))+" reference len="+strconv.Itoa(len(references)))
	}
	var results []db.Audio
	for i, line := range lines {
		var verse FAOutput
		err = json.Unmarshal(line, &verse)
		if err != nil {
			return log.Error(f.ctx, 500, err, `Error unmarshalling output file`)
		}
		fmt.Println(verse)
		var rec db.Audio
		rec.ScriptId = references[i]
		rec.BookId = file.BookId
		rec.ChapterNum = file.Chapter
		rec.AudioFile = file.Filename
		rec.BeginTS = math.Round(verse.AudioStart*1000.0) / 1000.0
		rec.EndTS = math.Round((verse.AudioStart+verse.Duration)*1000.0) / 1000.0
		results = append(results, rec)
	}
	status = f.conn.UpdateScriptFATimestamps(results)
	if status != nil {
		return status
	}
	return status
}
