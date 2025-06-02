package diff

import (
	"context"
	"database/sql"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/uroman"
	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/text/unicode/norm"
	"regexp"
	"strconv"
	"strings"
)

type Compare struct {
	ctx         context.Context
	user        string
	baseDataset string
	dataset     string
	baseDb      db.DBAdapter
	database    db.DBAdapter
	lang        string
	baseIdent   db.Ident
	compIdent   db.Ident
	testament   request.Testament
	settings    request.CompareSettings
	replacer    *strings.Replacer
	verseRm     *regexp.Regexp
	isLatin     sql.NullBool
	diffMatch   *diffmatchpatch.DiffMatchPatch
	results     []Pair
}

type Verse struct {
	scriptId   int
	bookId     string
	chapter    int
	chapterEnd int
	verse      string
	verseEnd   string
	scriptNum  string
	text       string
	uRoman     string
	beginTS    float64
	endTS      float64
}

func NewCompare(ctx context.Context, user string, baseDSet string, db db.DBAdapter,
	lang string, testament request.Testament, settings request.CompareSettings) Compare {
	var c Compare
	c.ctx = ctx
	c.user = user
	c.baseDataset = baseDSet
	c.dataset = db.Project
	c.database = db
	c.lang = lang
	c.testament = testament
	c.settings = settings
	c.replacer = c.cleanUpSetup()
	c.verseRm = regexp.MustCompile(`\{[0-9\-\,]+\}\s?`) // used by compareScriptLine
	c.isLatin.Valid = false
	c.diffMatch = diffmatchpatch.New()
	return c
}

func (c *Compare) Process() ([]Pair, string, string, *log.Status) {
	var records []Pair
	var fileMap string
	var languageISO string
	var status *log.Status
	var ident db.Ident
	ident, status = c.database.SelectIdent() // TextSource should be a parameter?
	if status != nil {
		return records, fileMap, languageISO, status
	}
	languageISO = ident.ASRLanguageISO
	if languageISO == "" {
		languageISO = ident.LanguageISO
	}
	c.baseDb, status = db.NewerDBAdapter(c.ctx, false, c.user, c.baseDataset)
	if status != nil {
		return records, fileMap, languageISO, status
	}
	status = uroman.EnsureUroman(c.baseDb, c.lang)
	if status != nil {
		return records, fileMap, languageISO, status
	}
	if ident.TextSource == request.TextScript {
		records, status = c.compareScriptLines(ident.TextSource)
	} else {
		records, status = c.compareVerses(ident.TextSource)
	}
	if status != nil {
		return records, fileMap, languageISO, status
	}
	fileMap, status = c.generateBookChapterFilenameMap()
	c.baseDb.Close()
	return records, fileMap, languageISO, status
}

func (c *Compare) compareVerses(textSource request.MediaType) ([]Pair, *log.Status) {
	var status *log.Status
	for _, bookId := range db.RequestedBooks(c.testament) {
		var chapInBook, _ = db.BookChapterMap[bookId]
		var chapter = 1
		for chapter <= chapInBook {
			var baseLines, compLines []Verse
			baseLines, status = c.process(c.baseDb, bookId, chapter, textSource)
			if status != nil {
				return c.results, status
			}
			compLines, status = c.process(c.database, bookId, chapter, textSource)
			if status != nil {
				return c.results, status
			}
			c.diff(baseLines, compLines)
			chapter++
		}
	}
	return c.results, status
}

func (c *Compare) process(conn db.DBAdapter, bookId string, chapterNum int, mediaType request.MediaType) ([]Verse, *log.Status) {
	var lines []Verse
	var status *log.Status
	scripts, status := conn.SelectScriptsByChapter(bookId, chapterNum)
	if status != nil {
		return lines, status
	}
	if !c.isLatin.Valid {
		c.SetIsLatin(scripts)
	}
	for _, script := range scripts {
		var vs Verse
		vs.scriptId = script.ScriptId
		vs.bookId = script.BookId
		vs.chapter = script.ChapterNum
		vs.chapterEnd = script.ChapterEnd
		vs.verse = script.VerseStr
		vs.verseEnd = script.VerseEnd
		if mediaType == request.TextScript {
			vs.scriptNum = script.ScriptNum
		}
		vs.text = script.ScriptText
		vs.uRoman = script.URoman
		vs.beginTS = script.ScriptBeginTS
		vs.endTS = script.ScriptEndTS
		lines = append(lines, vs)
	}
	//if ident.TextSource == request.TextScript {
	//	lines = c.consolidateScript(lines)
	//}
	lines = c.cleanUpVerses(lines, mediaType)
	return lines, status
}

// consolidateScript breaks a script into verses, it is needed for compare of script to USX or plain text
func (c *Compare) consolidateScript(verses []Verse) []Verse {
	const (
		begin = iota + 1
		inNum
		endNum
	)
	//var labels = []string{``, `BEGIN`, `INNUM`, `ENDNUM`}
	var results = make([]Verse, 0, len(verses))
	var sumInput = 0
	var sumOutput = 0
	var bookId = verses[0].bookId
	var chapter = verses[0].chapter
	var parts = make([]string, 0, 100)
	for _, rec := range verses {
		parts = append(parts, rec.text)
		sumInput += len(rec.text)
	}
	text := strings.Join(parts, ``)
	var verseNum = `0`
	var tmpNum []byte
	var index = 0
	var state = begin
	for index < len(text) {
		switch state {
		case begin:
			var part string
			search := text[index:]
			pos := strings.Index(search, `{`)
			if pos < 0 {
				part = search
				sumOutput += len(part)
			} else {
				part = search[:pos]
				state = inNum
				tmpNum = []byte{}
				sumOutput += len(part) + 1
			}
			verse := Verse{bookId: bookId, chapter: chapter, verse: verseNum, text: part}
			results = append(results, verse)
			index += len(part) + 1
		case inNum:
			char := text[index]
			if char >= '0' && char <= '9' {
				tmpNum = append(tmpNum, char)
				index++
				sumOutput += 1
			} else if char == '}' {
				verseNum = string(tmpNum)
				state = endNum
				index++
				sumOutput += 1
			} else {
				start := max(0, index-50)
				end := min(len(text)-1, index+50)
				log.Warn(c.ctx, bookId, chapter, verseNum, `Invalid char in {nn, expect n or } found `,
					string(char), ` in `, text[start:end])
				verseNum = string(tmpNum)
				state = begin
			}
		case endNum:
			char := text[index]
			peek := text[index+1]
			if (char == '_' || char == '-') && peek == '{' {
				state = inNum
				tmpNum = []byte(verseNum + "-")
				index += 2
				sumOutput += 2
			} else {
				state = begin
			}
		}
	}
	if sumInput != sumOutput {
		log.Warn(c.ctx, "Bug: Not all data processed by consolidateScript input:", sumInput, " output:", sumOutput)
	}
	return results
}

func (c *Compare) cleanUpSetup() *strings.Replacer {
	var replace []string
	cfg := c.settings
	if cfg.RemovePromptChars {
		replace = append(replace, "<<", "")
		replace = append(replace, ">>", "")
		replace = append(replace, "((", "")
		replace = append(replace, "†", "") // KDLTBL
		replace = append(replace, "[", "") // KDLTBL
		replace = append(replace, "]", "") // KDLTBL
		replace = append(replace, "<", "") // AAAMLT
		replace = append(replace, ">", "") // AAAMLT
	}
	if cfg.RemovePunctuation {
		replace = append(replace, ",", "")
		replace = append(replace, ";", "")
		replace = append(replace, ":", "")
		replace = append(replace, ".", "")
		replace = append(replace, "!", "")
		replace = append(replace, "?", "")
		replace = append(replace, "~", "")
	}
	if cfg.DoubleQuotes.Normalize {
		//replace = append(replace, "\u201C", "\u0022") // left quote
		//replace = append(replace, "\u201D", "\u0022") // right quote
		//replace = append(replace, "\u2033", "\u0022") // minutes or seconds
		replace = append(replace, "\u00AB", "\u0022") // <<
		//replace = append(replace, "\u00BB", "\u0022") // >>
		//replace = append(replace, "\u201E", "\u0022") // low 9 quote
	} else if cfg.DoubleQuotes.Remove {
		replace = append(replace, "\u0022", "") // ascii
		replace = append(replace, "\u201C", "") // left quote
		replace = append(replace, "\u201D", "") // right quote
		//replace = append(replace, "\u2033", "") // minutes or seconds
		replace = append(replace, "\u00AB", "") // <<
		replace = append(replace, "\u00BB", "") // >>
		//replace = append(replace, "\u201E", "") // low 9 quote
	}
	if cfg.Apostrophe.Normalize {
		replace = append(replace, "\uA78C", "'") // ? found in script text
		replace = append(replace, "\u2018", "'") // left
		replace = append(replace, "\u2019", "'") // right
		replace = append(replace, "\u02BC", "'") // modifier letter apos
		replace = append(replace, "\u2032", "'") // prime
		replace = append(replace, "\u00B4", "'") // acute accent (not really apos)
	} else if cfg.Apostrophe.Remove {
		replace = append(replace, "'", "")
		replace = append(replace, "\uA78C", "") // ? found in script text
		replace = append(replace, "\u2018", "") // left
		replace = append(replace, "\u2019", "") // right
		//replace = append(replace, "\u02BC", "") // modifier letter apos
		//replace = append(replace, "\u2032", "") // prime
		//replace = append(replace, "\u00B4", "") // acute accent (not really apos)
	}
	if cfg.Hyphen.Normalize {
		replace = append(replace, "\u2010", "\u002D") // hypen
		replace = append(replace, "\u2011", "\u002D") // non-breaking hyphen
		replace = append(replace, "\u2012", "\u002D") // figure dash
		replace = append(replace, "\u2013", "\u002D") // en dash
		replace = append(replace, "\u2014", "\u002D") // em dash
		replace = append(replace, "\u2015", "\u002D") // horizontal bar
		replace = append(replace, "\uFE58", "\u002D") // small em dash
		replace = append(replace, "\uFE62", "\u002D") // small en dash
		replace = append(replace, "\uFE63", "\u002D") // small hyphen minus
		replace = append(replace, "\uFF0D", "\u002D") // fullwidth hypen-minus
	} else if cfg.Hyphen.Remove {
		replace = append(replace, "\u002D", "") // hypen
		replace = append(replace, "\u2010", "") // hypen
		replace = append(replace, "\u2011", "") // non-breaking hyphen
		replace = append(replace, "\u2012", "") // figure dash
		replace = append(replace, "\u2013", "") // en dash
		replace = append(replace, "\u2014", "") // em dash
		replace = append(replace, "\u2015", "") // horizontal bar
		replace = append(replace, "\uFE58", "") // small em dash
		replace = append(replace, "\uFE62", "") // small en dash
		replace = append(replace, "\uFE63", "") // small hyphen minus
		replace = append(replace, "\uFF0D", "") // fullwidth hypen-minus
	}
	var replacer *strings.Replacer
	if len(replace) > 0 {
		replacer = strings.NewReplacer(replace...)
	}
	return replacer
}

func (c *Compare) cleanUpVerses(verses []Verse, mediaType request.MediaType) []Verse {
	for i := range verses {
		verses[i].text = c.cleanup(verses[i].text)
		verses[i].uRoman = c.cleanup(verses[i].uRoman)
		if mediaType == request.TextScript {
			verses[i].text = c.verseRm.ReplaceAllString(verses[i].text, ``)
		}
	}
	return verses
}

func (c *Compare) cleanup(text string) string {
	if c.replacer != nil {
		text = c.replacer.Replace(text)
	}
	cfg := c.settings
	if cfg.LowerCase {
		text = strings.ToLower(text)
	}
	// https://unicode.org/reports/tr15/  Normalization Doc
	//if cfg.DiacriticalMarks.Normalize {
	if cfg.DiacriticalMarks.NormalizeNFC {
		text = norm.NFC.String(text)
	} else if cfg.DiacriticalMarks.NormalizeNFD {
		text = norm.NFD.String(text)
	} else if cfg.DiacriticalMarks.NormalizeNFKC {
		text = norm.NFKC.String(text)
	} else if cfg.DiacriticalMarks.NormalizeNFKD {
		text = norm.NFKD.String(text)
	} else if cfg.DiacriticalMarks.Remove {
		text = norm.NFKD.String(text)
		var filtered = make([]rune, 0, 300)
		for _, ch := range text {
			if ch < '\u0300' || ch > '\u036F' {
				filtered = append(filtered, ch)
			}
		}
		text = string(filtered)
	}
	return text
}

/* This diff method assumes one chapter at a time */
func (c *Compare) diff(baseVS []Verse, compVS []Verse) {
	var didMatch = make(map[string]bool)
	// Put the second data in a map
	var verse2Map = make(map[string]Verse)
	for _, vs2 := range compVS {
		verse2Map[vs2.verse] = vs2
	}
	// combine the verse2 to verse1 that match
	var p Pair
	for _, vs1 := range baseVS {
		vs2, ok := verse2Map[vs1.verse]
		if ok {
			didMatch[vs1.verse] = true
			p = NewPair(&vs1, &vs2)
		} else {
			p = NewPair(&vs1, nil)
		}
		c.diffPair(p)
	}
	// pick up any verse2 that did not match verse1
	for _, vs2 := range compVS {
		_, ok := didMatch[vs2.verse]
		if !ok {
			p = NewPair(nil, &vs2)
			c.diffPair(p)
		}
	}
}

func (c *Compare) diffPair(pair Pair) {
	baseText, compText := pair.Text(c.isLatin)
	if len(baseText) > 0 || len(compText) > 0 {
		baseText = strings.TrimSpace(baseText)
		compText = strings.TrimSpace(compText)
		diffs := c.diffMatch.DiffMain(baseText, compText, false)
		pair.Diffs = c.diffMatch.DiffCleanupMerge(diffs) // required for measure to compute largest
		if !c.isMatch(pair.Diffs) {
			pair.HTML = c.diffMatch.DiffPrettyHtml(pair.Diffs)
			c.results = append(c.results, pair)
		}
	}
}

func (c *Compare) isMatch(diffs []diffmatchpatch.Diff) bool {
	for _, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffInsert || diff.Type == diffmatchpatch.DiffDelete {
			if len(strings.TrimSpace(diff.Text)) > 0 {
				return false
			}
		}
	}
	return true
}

// ensureClean is deprecated, we are using diff_clean_merge
func (c *Compare) ensureClean(diffs []diffmatchpatch.Diff) {
	var prior diffmatchpatch.Operation
	for _, diff := range diffs {
		if diff.Type == prior {
			log.Warn(c.ctx, "There are inserts and deletes together, must use diff_clean_merge")
		}
		prior = diff.Type
	}
}

// generateBookChapterFilenameMap could be eliminated.  The filenames are now passed in []Pair slice
func (c *Compare) generateBookChapterFilenameMap() (string, *log.Status) {
	chapters, status := c.database.SelectBookChapterFilename()
	if status != nil {
		return "", status
	}
	var result []string
	result = append(result, "let fileMap = {\n")
	for i, ch := range chapters {
		key := ch.BookId + strconv.Itoa(ch.ChapterNum)
		result = append(result, "\t'"+key+"': '"+ch.AudioFile+"'")
		if i < len(chapters)-1 {
			result = append(result, ",\n")
		} else {
			result = append(result, "};\n")
		}
	}
	return strings.Join(result, ""), status
}

func (c *Compare) SetIsLatin(records []db.Script) {
	if !c.isLatin.Valid {
		var numChars = 0
		var numLatin = 0
		for _, rec := range records {
			for _, ch := range []rune(rec.ScriptText) {
				numChars++
				if ch <= '\u024F' { // Upper Bound of Latin Extended-B
					numLatin++
				}
			}
		}
		pct := float64(numLatin) / float64(numChars)
		c.isLatin.Valid = true
		c.isLatin.Bool = pct > 0.9
	}
}
