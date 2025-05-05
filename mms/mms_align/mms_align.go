package mms_align

import (
	"context"
	"encoding/json"
	"github.com/divan/num2words"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/generic"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/ffmpeg"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/stdio_exec"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/uroman"
	"golang.org/x/text/unicode/norm"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type MMSAlign_Input struct {
	AudioFile string   `json:"audio_file"`
	NormWords []string `json:"words"`
}

type Word struct {
	verseStr string
	scriptId int64
	wordId   int64
	wordSeq  int
	word     string
	uroman   string
}

type MMSAlign struct {
	ctx      context.Context
	conn     db.DBAdapter // This database adapter must contain the text to be processed
	lang     string
	sttLang  string
	tempDir  string
	uroman   stdio_exec.StdioExec
	mmsAlign stdio_exec.StdioExec
}

func NewMMSAlign(ctx context.Context, conn db.DBAdapter, lang string, sttLang string) MMSAlign {
	var m MMSAlign
	m.ctx = ctx
	m.conn = conn
	m.lang = lang
	m.sttLang = sttLang
	return m
}

// ProcessFiles will perform Forced Alignment on these files
func (m *MMSAlign) ProcessFiles(files []input.InputFile) *log.Status {
	var status *log.Status
	var err error
	m.tempDir, err = os.MkdirTemp(os.Getenv(`FCBH_DATASET_TMP`), "mms_fa_")
	if err != nil {
		return log.Error(m.ctx, 500, err, `Error creating temp dir`)
	}
	defer os.RemoveAll(m.tempDir)
	m.uroman, status = stdio_exec.NewStdioExec(m.ctx, os.Getenv(`FCBH_MMS_FA_PYTHON`), uroman.ScriptPath(), "-l", m.lang)
	if status != nil {
		return status
	}
	defer m.uroman.Close()
	pythonScript := filepath.Join(os.Getenv("GOPROJ"), "mms/mms_align.py")
	m.mmsAlign, status = stdio_exec.NewStdioExec(m.ctx, os.Getenv(`FCBH_MMS_FA_PYTHON`), pythonScript)
	if status != nil {
		return status
	}
	defer m.mmsAlign.Close()
	for _, file := range files {
		log.Info(m.ctx, "MMS Align", file.BookId, file.Chapter)
		status = m.processFile(file)
		if status != nil {
			return status
		}
	}
	return status
}

// processFile will process one audio file through mms forced alignment
func (m *MMSAlign) processFile(file input.InputFile) *log.Status {
	var status *log.Status
	var faInput MMSAlign_Input
	faInput.AudioFile, status = ffmpeg.ConvertMp3ToWav(m.ctx, m.tempDir, file.FilePath())
	if status != nil {
		return status
	}
	var wordList []Word
	faInput.NormWords, wordList, status = m.prepareText(m.lang, file.BookId, file.Chapter)
	if status != nil {
		return status
	}
	content, err := json.Marshal(faInput)
	if err != nil {
		return log.Error(m.ctx, 500, err, `Error marshalling json`)
	}
	// development
	//err2 := os.WriteFile("engweb_fa_inp.json", content, 0644)
	//if err2 != nil {
	//	panic(err2)
	//}
	response, status := m.mmsAlign.Process(string(content))
	if status != nil {
		return status
	}
	// development
	//fmt.Println(len(wordList)) // temp
	//err = os.WriteFile("engweb_fa_out.json", []byte(response), 0644)
	//if err != nil {
	//	panic(err)
	//}
	status = m.processPyOutput(file, wordList, response)
	return status
}

func (m *MMSAlign) prepareText(lang string, bookId string, chapter int) ([]string, []Word, *log.Status) {
	var textList []string
	var refList []Word
	var dbWords, status = m.conn.SelectWordsByBookChapter(bookId, chapter)
	if status != nil {
		return textList, refList, status
	}
	for _, wd := range dbWords {
		var ref Word
		ref.verseStr = wd.VerseStr
		ref.scriptId = int64(wd.ScriptId)
		ref.wordId = int64(wd.WordId)
		ref.wordSeq = wd.WordSeq
		ref.word = norm.NFC.String(wd.Word)
		uRoman, status2 := m.uroman.Process(ref.word)
		if status2 != nil {
			return textList, refList, status2
		}
		word := m.convertNum2Words(uRoman) // This does NOT handle isolated digits, only whole word numbers
		word = m.normalizeURoman(word)
		ref.uroman = word
		refList = append(refList, ref)
		textList = append(textList, word)
	}
	if len(textList) != len(refList) {
		status = log.ErrorNoErr(m.ctx, 500, "mms_align.prepareText created lists of different sizes", len(textList), len(refList))
	}
	return textList, refList, status
}

func (m *MMSAlign) convertNum2Words(text string) string {
	for _, ch := range []rune(text) {
		if !unicode.IsDigit(ch) && ch != '.' && ch != ',' && ch != '-' {
			return text
		}
	}
	num, _ := strconv.Atoi(text)
	return num2words.Convert(num)
}

// normalizeURoman is taken precisely from torchaudio documentation
func (m *MMSAlign) normalizeURoman(text string) string {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, "\u2019", "'")
	re1 := regexp.MustCompile("[^a-z' ]")
	text2 := re1.ReplaceAllString(text, " ")
	re2 := regexp.MustCompile(" +")
	text2 = re2.ReplaceAllString(text2, " ")
	// This line is only to be used when doing one word at a time
	// It is needed to maintain the word alignment.
	text2 = strings.ReplaceAll(text2, " ", "")
	if text2 != text {
		log.Warn(m.ctx, "Changed:", text, " To:", text2)
	}
	return strings.TrimSpace(text2)
}

type MMSAlignResult struct {
	Ratio      float64        `json:"ratio"`
	Dictionary map[string]int `json:"dictionary"`
	Tokens     [][][]float64  `json:"tokens"`
}

func (m *MMSAlign) processPyOutput(file input.InputFile, wordRefs []Word, response string) *log.Status {
	var status *log.Status
	var mmsAlign MMSAlignResult
	err := json.Unmarshal([]byte(response), &mmsAlign)
	if err != nil {
		return log.Error(m.ctx, 500, err, `Error unmarshalling json`)
	}
	var tokenDict = make(map[int]rune)
	for chr, token := range mmsAlign.Dictionary {
		tokenDict[token] = []rune(chr)[0]
	}
	var ok bool
	var faWords [][]generic.Char
	for _, wd := range mmsAlign.Tokens {
		var word []generic.Char
		for _, ch := range wd {
			var char generic.Char
			token := int(ch[0])
			char.Start = ch[1] * mmsAlign.Ratio
			char.End = ch[2] * mmsAlign.Ratio
			char.Score = ch[3]
			char.Uroman, ok = tokenDict[token]
			if !ok {
				log.Warn(m.ctx, "Character not found in tokenDict", token)
			}
			word = append(word, char)
		}
		faWords = append(faWords, word)
	}
	if len(faWords) != len(wordRefs) {
		return log.ErrorNoErr(m.ctx, 400, "Num words input to mms_align:", len(wordRefs), ", num timestamps returned:", len(faWords))
	}
	var words []db.Audio
	for i, ref := range wordRefs {
		var word db.Audio
		word.BookId = file.BookId
		word.ChapterNum = file.Chapter
		word.AudioFile = file.Filename
		word.ScriptId = ref.scriptId
		word.WordId = ref.wordId // because hypenated words were split, multiple words can have the same wordId
		word.WordSeq = ref.wordSeq
		word.Text = ref.word
		word.Uroman = ref.uroman
		faWd := faWords[i]
		if len(faWd) > 0 {
			word.BeginTS = faWd[0].Start
			word.EndTS = faWd[len(faWd)-1].End
		}
		uromanChars := []rune(ref.uroman)
		for j, ch := range faWd {
			word.FAScore += ch.Score
			faWd[j].Seq = j
			if faWd[j].Uroman != uromanChars[j] {
				log.ErrorNoErr(m.ctx, 500, "Norm", ref.uroman, "does not match")
			}
		}
		if len(faWd) > 0 {
			word.FAScore = word.FAScore / float64(len(faWd))
		}
		word.Chars = faWd
		words = append(words, word)
	}
	var wordsByLine [][]db.Audio
	wordsByLine = m.groupByLine(words)
	var verses []db.Audio
	verses = m.summarizeByVerse(wordsByLine)
	verses = m.midPoint(verses)
	verses[len(verses)-1].EndTS, status = ffmpeg.GetAudioDuration(m.ctx, file.Directory, file.Filename)
	if status != nil {
		return status
	}
	status = m.conn.UpdateScriptFATimestamps(verses)
	if status != nil {
		return status
	}
	status = m.conn.UpdateWordFATimestamps(words)
	if status != nil {
		return status
	}
	status = m.conn.InsertAudioChars(words)
	return status
}

func (m *MMSAlign) groupByLine(words []db.Audio) [][]db.Audio {
	var result [][]db.Audio
	if len(words) == 0 {
		return result
	}
	currWd := words[0].ScriptId
	start := 0
	for i, wd := range words {
		if wd.ScriptId != currWd {
			currWd = wd.ScriptId
			oneLine := make([]db.Audio, i-start)
			copy(oneLine, words[start:i])
			result = append(result, oneLine)
			start = i
		}
	}
	if start != len(words) {
		lastLine := make([]db.Audio, len(words)-start)
		copy(lastLine, words[start:])
		result = append(result, lastLine)
	}
	return result
}

func (m *MMSAlign) summarizeByVerse(chapter [][]db.Audio) []db.Audio {
	var result []db.Audio
	for _, verse := range chapter {
		var vs = verse[0]
		vs.EndTS = verse[len(verse)-1].EndTS
		var text []string
		var uroman []string
		var scores []float64
		for _, word := range verse {
			text = append(text, word.Text)
			uroman = append(uroman, word.Uroman)
			scores = append(scores, word.FAScore)
		}
		vs.Text = strings.Join(text, " ")
		vs.Uroman = strings.Join(uroman, " ")
		vs.FAScore = m.average(scores, 5)
		result = append(result, vs)
	}
	return result
}

func (m *MMSAlign) average(scores []float64, precision int) float64 {
	var sum float64
	for _, scr := range scores {
		sum += scr
	}
	avg := sum / float64(len(scores))
	pow := math.Pow10(precision)
	result := math.Round(avg*pow) / pow
	return result
}

// midPoint eliminates time gaps between the end of one verse and the beginning of the next.
func (m *MMSAlign) midPoint(parts []db.Audio) []db.Audio {
	for i := range parts {
		if i == 0 {
			parts[0].BeginTS = 0.0
		} else {
			midPoint := (parts[i].BeginTS + parts[i-1].EndTS) / 2.0
			parts[i].BeginTS = midPoint
			parts[i-1].EndTS = midPoint
		}
	}
	return parts
}
