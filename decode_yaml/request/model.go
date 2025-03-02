package request

type Request struct {
	IsNew         bool          `yaml:"is_new"`
	DatasetName   string        `yaml:"dataset_name"`
	BibleId       string        `yaml:"bible_id"`
	Username      string        `yaml:"username"`
	Email         string        `yaml:"email"`
	AltLanguage   string        `yaml:"alt_language,omitempty"`
	Output        Output        `yaml:"output,omitempty"`
	Testament     Testament     `yaml:"testament,omitempty"`
	Database      Database      `yaml:"database,omitempty"`
	AudioData     AudioData     `yaml:"audio_data,omitempty"`
	TextData      TextData      `yaml:"text_data,omitempty"`
	Timestamps    Timestamps    `yaml:"timestamps,omitempty"`
	SpeechToText  SpeechToText  `yaml:"speech_to_text,omitempty"`
	Detail        Detail        `yaml:"detail,omitempty"`
	AudioEncoding AudioEncoding `yaml:"audio_encoding,omitempty"`
	TextEncoding  TextEncoding  `yaml:"text_encoding,omitempty"`
	AudioProof    AudioProof    `yaml:"audio_proof,omitempty"`
	Compare       Compare       `yaml:"compare,omitempty"`
	UpdateDBP     UpdateDBP     `yaml:"update_dbp,omitempty"`
}

// GetTestUser is used for testing when there is no full request object.
func GetTestUser() string {
	return `GaryNTest`
}

type Output struct {
	Directory string `yaml:"directory"`
	CSV       bool   `yaml:"csv,omitempty"`
	JSON      bool   `yaml:"json,omitempty"`
	Sqlite    bool   `yaml:"sqlite,omitempty"`
}

type Testament struct {
	NT      bool     `yaml:"nt,omitempty"`
	NTBooks []string `yaml:"nt_books,omitempty"`
	OT      bool     `yaml:"ot,omitempty"`
	OTBooks []string `yaml:"ot_books,omitempty"`
	otMap   map[string]bool
	ntMap   map[string]bool
}

func (t *Testament) BuildBookMaps() {
	t.otMap = make(map[string]bool)
	for _, book := range t.OTBooks {
		t.otMap[book] = true
	}
	t.ntMap = make(map[string]bool)
	for _, book := range t.NTBooks {
		t.ntMap[book] = true
	}
}

func (t *Testament) Has(ttype string, bookId string) bool {
	if ttype == `NT` {
		return t.HasNT(bookId)
	} else {
		return t.HasOT(bookId)
	}
}

func (t *Testament) HasOT(bookId string) bool {
	if t.OT {
		return true
	}
	_, ok := t.otMap[bookId]
	return ok
}

func (t *Testament) HasNT(bookId string) bool {
	if t.NT {
		return true
	}
	_, ok := t.ntMap[bookId]
	return ok
}

type Database struct {
	AWSS3 string `yaml:"aws_s3,omitempty"`
}

type AudioData struct {
	BibleBrain BibleBrainAudio `yaml:"bible_brain,omitempty"`
	File       string          `yaml:"file,omitempty"`
	AWSS3      string          `yaml:"aws_s3,omitempty"`
	POST       string          `yaml:"post,omitempty"`
	NoAudio    bool            `yaml:"no_audio,omitempty"`
}

func (a AudioData) AnyBibleBrain() bool {
	return a.BibleBrain.MP3_64 || a.BibleBrain.MP3_16 || a.BibleBrain.OPUS
}

type BibleBrainAudio struct {
	MP3_64 bool `yaml:"mp3_64,omitempty"`
	MP3_16 bool `yaml:"mp3_16,omitempty"`
	OPUS   bool `yaml:"opus,omitempty"`
}

func (b BibleBrainAudio) AudioType() (string, string) {
	var result string
	var codec string
	if b.MP3_64 {
		result = `MP3`
		codec = `64kbps`
	} else if b.MP3_16 {
		result = `MP3`
		codec = `16kbps`
	} else if b.OPUS {
		result = `OPUS`
		codec = ``
	}
	return result, codec
}

type TextData struct {
	BibleBrain BibleBrainText `yaml:"bible_brain,omitempty"`
	File       string         `yaml:"file,omitempty"`
	AWSS3      string         `yaml:"aws_s3,omitempty"`
	POST       string         `yaml:"post,omitempty"`
	NoText     bool           `yaml:"no_text,omitempty"`
}

func (t TextData) AnyBibleBrain() bool {
	return t.BibleBrain.TextUSXEdit || t.BibleBrain.TextPlainEdit || t.BibleBrain.TextPlain
}

type BibleBrainText struct {
	TextUSXEdit   bool `yaml:"text_usx_edit,omitempty"`
	TextPlainEdit bool `yaml:"text_plain_edit,omitempty"`
	TextPlain     bool `yaml:"text_plain,omitempty"`
}

type MediaType string

const (
	Audio         MediaType = "audio"
	AudioDrama    MediaType = "audio_drama"
	TextUSXEdit   MediaType = "text_usx_edit"
	TextPlainEdit MediaType = "text_plain_edit"
	TextPlain     MediaType = "text_plain"
	TextScript    MediaType = "text_script"
	TextCSV       MediaType = "text_csv"
	TextSTT       MediaType = "text_stt"
	TextNone      MediaType = ""
)

func (b BibleBrainText) TextType() MediaType {
	var result MediaType
	if b.TextUSXEdit {
		result = TextUSXEdit
	} else if b.TextPlainEdit {
		result = TextPlainEdit
	} else if b.TextPlain {
		result = TextPlain
	} else {
		result = TextNone
	}
	return result
}

func (t MediaType) IsFrom(ttype string) bool {
	var result = false
	switch t {
	case TextUSXEdit:
		result = ttype == `text_usx`
	case TextPlainEdit:
		result = ttype == `text_plain`
	case TextPlain:
		result = ttype == `text_plain`
	case TextScript:
		result = ttype == `text_script`
	case TextNone:
		result = ttype == `text_none`
	}
	return result
}

type SpeechToText struct {
	MMS            bool    `yaml:"mms_asr,omitempty"`
	Whisper        Whisper `yaml:"whisper,omitempty"`
	NoSpeechToText bool    `yaml:"no_speech_to_text,omitempty"`
}

type Whisper struct {
	Model WhisperModel `yaml:"model,omitempty"`
}
type WhisperModel struct {
	Large  bool `yaml:"large,omitempty"`
	Medium bool `yaml:"medium,omitempty"`
	Small  bool `yaml:"small,omitempty"`
	Base   bool `yaml:"base,omitempty"`
	Tiny   bool `yaml:"tiny,omitempty"`
}

func (w WhisperModel) String() string {
	var result string
	if w.Large {
		result = `large`
	} else if w.Medium {
		result = `medium`
	} else if w.Small {
		result = `small`
	} else if w.Base {
		result = `base`
	} else if w.Tiny {
		result = `tiny`
	}
	return result
}

type Detail struct {
	Lines bool `yaml:"lines,omitempty"`
	Words bool `yaml:"words,omitempty"`
}

type Timestamps struct {
	BibleBrain   bool `yaml:"bible_brain,omitempty"`
	Aeneas       bool `yaml:"aeneas,omitempty"`
	TSBucket     bool `yaml:"ts_bucket,omitempty"`
	MMSFAVerse   bool `yaml:"mms_fa_verse,omitempty"`
	MMSAlign     bool `yaml:"mms_align,omitempty"`
	NoTimestamps bool `yaml:"no_timestamps,omitempty"`
}

type AudioEncoding struct {
	MFCC       bool `yaml:"mfcc,omitempty"`
	NoEncoding bool `yaml:"no_encoding,omitempty"`
}

type TextEncoding struct {
	FastText   bool `yaml:"fast_text,omitempty"`
	NoEncoding bool `yaml:"no_encoding,omitempty"`
}

type AudioProof struct {
	HTMLReport  bool   `yaml:"html_report,omitempty"`
	BaseDataset string `yaml:"base_dataset,omitempty"`
}

type Compare struct {
	HTMLReport      bool            `yaml:"html_report,omitempty"`
	BaseDataset     string          `yaml:"base_dataset,omitempty"`
	CompareSettings CompareSettings `yaml:"compare_settings,omitempty"`
}

type CompareSettings struct {
	LowerCase         bool              `yaml:"lower_case,omitempty"`
	RemovePromptChars bool              `yaml:"remove_prompt_chars,omitempty"`
	RemovePunctuation bool              `yaml:"remove_punctuation,omitempty"`
	DoubleQuotes      CompareChoice     `yaml:"double_quotes,omitempty"`
	Apostrophe        CompareChoice     `yaml:"apostrophe,omitempty"`
	Hyphen            CompareChoice     `yaml:"hyphen,omitempty"`
	DiacriticalMarks  DiacriticalChoice `yaml:"diacritical_marks,omitempty"`
}

type CompareChoice struct {
	Remove    bool `yaml:"remove,omitempty"`
	Normalize bool `yaml:"normalize,omitempty"`
}

type DiacriticalChoice struct {
	Remove        bool `yaml:"remove,omitempty"`
	NormalizeNFC  bool `yaml:"normalize_nfc,omitempty"`
	NormalizeNFD  bool `yaml:"normalize_nfd,omitempty"`
	NormalizeNFKC bool `yaml:"normalize_nfkc,omitempty"`
	NormalizeNFKD bool `yaml:"normalize_nfkd,omitempty"`
}

type UpdateDBP struct {
	Timestamps []string `yaml:"timestamps,omitempty"`
}
