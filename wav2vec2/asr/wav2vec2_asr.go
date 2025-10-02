package asr

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/stdio_exec"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/uroman"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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
	status = a.conn.UpdateASRLanguage(lang)
	if status != nil {
		return status
	}
	pythonScript := filepath.Join(os.Getenv("GOPROJ"), "wav2vec2/asr/wav2vec2_asr.py")
	// The -u option ensures the stdio is unbuffered.
	a.asrPy, status = stdio_exec.NewStdioExec(a.ctx, os.Getenv(`FCBH_MMS_ASR_PYTHON`), pythonScript, "-u", lang)
	if status != nil {
		return status
	}
	defer a.asrPy.Close()
	a.uroman, status = stdio_exec.NewStdioExec(a.ctx, os.Getenv(`FCBH_MMS_FA_PYTHON`), uroman.ScriptPath(), "-l", a.lang)
	if status != nil {
		return status
	}
	defer a.uroman.Close()
	project := a.conn.Project
	if strings.HasSuffix(project, "_audio") {
		project = project[:len(project)-len("_audio")]
	}
	sampleDBPath := filepath.Join(os.Getenv(`FCBH_DATASET_TMP`), project+`.db`)
	sampleDB := db.NewDBAdapter(a.ctx, sampleDBPath)
	for _, file := range files {
		log.Info(a.ctx, "Wav2Vec2 ASR", file.BookId, file.Chapter)
		status = a.processFile(file, sampleDB)
		if status != nil {
			return status
		}
	}
	return status
}

type verseRec struct {
	scriptId int64
	verseStr string
	verseNum int //(abc problems)
	words    []wordRec
}
type wordRec struct {
	wordId  int64
	wordNum int
	text    string
	input   []byte
	idx     int
}

func (a *Wav2Vec2ASR) processFile(file input.InputFile, sampleDB db.DBAdapter) *log.Status {
	var audioFiles []db.Audio
	verses, status := a.selectSample(sampleDB, file.BookId, file.Chapter)
	if status != nil {
		return status
	}
	for _, verse := range verses {
		var wordList []string
		for _, word := range verse.words {
			response, status1 := a.asrPy.ProcessBytes(word.input)
			if status1 != nil {
				return status1
			}
			log.Info(a.ctx, file.BookId, file.Chapter, verse.verseStr, word.wordNum, response)
			wordList = append(wordList, response)
		}
		verseStr := strings.Join(wordList, " ")
		var audioFile db.Audio
		audioFile.ScriptId = verse.scriptId
		audioFile.Text = verseStr
		uRoman, status2 := a.uroman.Process(verseStr)
		if status2 != nil {
			return status2
		}
		audioFile.Uroman = uRoman
		audioFiles = append(audioFiles, audioFile)
	}
	//log.Debug(a.ctx, "Finished ASR", file.BookId, file.Chapter)
	var recCount int
	recCount, status = a.conn.UpdateScriptText(audioFiles)
	if recCount != len(audioFiles) {
		log.Warn(a.ctx, "Timestamp update counts need investigation", recCount, len(audioFiles))
	}
	return status
}

func (a *Wav2Vec2ASR) selectSample(db db.DBAdapter, bookId string, chapter int) ([]verseRec, *log.Status) {
	var verses []verseRec
	key := bookId + " " + strconv.Itoa(chapter)
	query := `SELECT idx, script_id, word_id, input_values, text, reference FROM samples WHERE reference like '` + key + `:%'`
	rows, err := db.DB.Query(query)
	if err != nil {
		return verses, log.Error(a.ctx, 500, err, `Error reading rows in selectSample`)
	}
	defer rows.Close()
	verseMap := make(map[string]*verseRec)
	verseOrder := []string{} // Track order of verses
	for rows.Next() {
		var word wordRec
		var reference string
		var scriptId int64
		err = rows.Scan(&word.idx, &scriptId, &word.wordId, &word.input, &word.text, &reference)
		if err != nil {
			return verses, log.Error(a.ctx, 500, err, `Error scanning in selectSample`)
		}
		parts := strings.Split(reference, ":")
		if len(parts) != 2 {
			return verses, log.Error(a.ctx, 500, err, `Error scanning in selectSample`)
		}
		pieces := strings.Split(parts[1], ".")
		if len(pieces) != 2 {
			return verses, log.Error(a.ctx, 500, err, `Error scanning in selectSample`)
		}
		verseStr := pieces[0]
		word.wordNum, err = strconv.Atoi(pieces[1])
		if err != nil {
			return verses, log.Error(a.ctx, 500, err, `Error scanning in selectSample`)
		}
		verse, exists := verseMap[verseStr]
		if !exists {
			verse = &verseRec{
				scriptId: scriptId,
				verseStr: verseStr,
				words:    []wordRec{},
			}
			verseMap[verseStr] = verse
			verseOrder = append(verseOrder, verseStr) // Track order
		}
		verse.words = append(verse.words, word)
	}
	err = rows.Err()
	if err != nil {
		return verses, log.Error(a.ctx, 500, err, `Error at end of rows in ReadingScriptByChapter`)
	}
	verses = make([]verseRec, 0, len(verseMap))
	for _, verseStr := range verseOrder {
		verse := verseMap[verseStr]
		sort.Slice(verse.words, func(i, j int) bool {
			return verse.words[i].wordNum < verse.words[j].wordNum
		})
		verses = append(verses, *verse)
	}
	sort.Slice(verses, func(i, j int) bool {
		numI, _ := strconv.Atoi(verses[i].verseStr)
		numJ, _ := strconv.Atoi(verses[j].verseStr)
		return numI < numJ
	})
	return verses, nil
}
