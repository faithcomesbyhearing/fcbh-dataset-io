package db

import (
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/generic"
)

type Ident struct {
	DatasetId    int64
	BibleId      string
	AudioOTId    string
	AudioNTId    string
	TextOTId     string
	TextNTId     string
	TextSource   request.MediaType
	LanguageISO  string
	VersionCode  string
	LanguageId   int
	RolvId       int
	Alphabet     string
	LanguageName string
	VersionName  string
}

type Script struct {
	ScriptId      int
	DatasetId     int
	BookId        string
	ChapterNum    int
	ChapterEnd    int
	VerseStr      string
	VerseEnd      string
	VerseNum      int
	AudioFile     string
	ScriptNum     string
	UsfmStyle     string
	Person        string
	Actor         string
	ScriptText    string
	URoman        string
	ScriptTexts   []string
	ScriptBeginTS float64
	ScriptEndTS   float64
}

type Word struct {
	VerseStr    string
	WordId      int
	ScriptId    int
	WordSeq     int
	VerseNum    int
	TType       string
	Word        string
	WordBeginTS float64
	WordEndTS   float64
	FAScore     float64
	WordEncoded []float64
}

type Timestamp struct {
	Id        int
	VerseStr  string
	AudioFile string
	Text      string
	BeginTS   float64
	EndTS     float64
}

type MFCC struct {
	Id   int
	Rows int
	Cols int
	MFCC [][]float32
}

type Audio struct {
	WordId          int64          `json:"word_id,omitempty"` // Used only for Word table
	ScriptId        int64          `json:"script_id"`
	BookId          string         `json:"book_id"`
	ChapterNum      int            `json:"chapter_num"`
	ChapterEnd      int            `json:"chapter_end"`
	VerseStr        string         `json:"verse_str"`
	VerseEnd        string         `json:"verse_end"`
	VerseSeq        int            `json:"verse_seq"`
	WordSeq         int            `json:"word_seq"` // Used by Words, not Verses
	BeginTS         float64        `json:"begin_ts"`
	EndTS           float64        `json:"end_ts"`
	FAScore         float64        `json:"fa_score"`
	Uroman          string         `json:"uroman"` // Is this useful?
	Text            string         `json:"text"`
	Chars           []generic.Char `json:"-"`
	AudioFile       string         `json:"audio_file"`
	ScriptBeginTS   float64        `json:"script_begin_ts"` // Contains script TS when it is a word record
	ScriptEndTS     float64        `json:"script_end_ts"`   // Contains script TS when it is a word record
	ScriptFAScore   float64        `json:"script_fa_score"` // Contains script score when it is a word record
	AudioChapterWav string         `json:"-"`               // Transient
	AudioVerseWav   string         `json:"-"`               // Transient
}

/*
type AlignChar struct {
	ScriptId     int64
	BookId       string
	ChapterNum   int
	VerseStr     string
	WordId       int64
	WordSeq      int
	Word         string
	CharId       int64
	CharSeq      int
	CharNorm     rune
	CharUroman   rune
	BeginTS      float64
	EndTS        float64
	FAScore      float64
	Duration     float64 // might not be needed
	Silence      float64
	SilencePos   int
	ScoreError   int
	DurationLong int
	SilenceLong  int
}
*/
