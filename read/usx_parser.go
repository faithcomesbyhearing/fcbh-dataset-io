package read

import (
	"context"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/safe"
)

type Stack []string

var hasStyle = map[string]bool{
	`book`: true, `para`: true, `char`: true, `cell`: true, `ms`: true, `note`: true, `sidebar`: true, `figure`: true}

//var numericPattern = regexp.MustCompile(`^\d+`)

type USXParser struct {
	ctx  context.Context
	conn db.DBAdapter
}

func NewUSXParser(conn db.DBAdapter) USXParser {
	var p USXParser
	p.ctx = conn.Ctx
	p.conn = conn
	return p
}

func (p *USXParser) ProcessFiles(inputFiles []input.InputFile) *log.Status {
	var status *log.Status
	for _, file := range inputFiles {
		filename := filepath.Join(file.Directory, file.Filename)
		var records []db.Script
		var titles titleDesc
		records, titles, status = p.decode(p.ctx, filename) // Also edits out non-script elements
		if status != nil {
			return status
		}
		records = p.addChapterHeading(records, titles)
		records = p.correctScriptNum(records)
		status = p.conn.InsertScripts(records)
		if status != nil {
			return status
		}
	}
	return status
}

func (p *USXParser) decode(ctx context.Context, filename string) ([]db.Script, titleDesc, *log.Status) {
	var records []db.Script
	var titles titleDesc
	var status *log.Status
	xmlFile, err := os.Open(filename)
	if err != nil {
		return records, titles, log.Error(ctx, 500, err, "USXParser could not open USX File.")
	}
	defer xmlFile.Close()
	var stack Stack
	var rec db.Script
	var tagName string
	var chapterNum = 1
	var scriptNum = 0
	var verseNum int
	var verseStr = `0`
	var usfmStyle string
	decoder := xml.NewDecoder(xmlFile)
	for {
		var token xml.Token
		token, err = decoder.Token()
		if err == io.EOF {
			break // End of file
		}
		if err != nil {
			return records, titles, log.Error(ctx, 500, err, "Error parsing USX.")
		}
		switch se := token.(type) {
		//[StartElement], [EndElement], [CharData], [Comment], [ProcInst], [Directive].
		case xml.StartElement:
			tagName = se.Name.Local
			if tagName == `book` {
				rec.BookId = p.findAttr(se, `code`)
			} else if tagName == `chapter` {
				chapterNum = p.findIntAttr(se, `number`)
			} else if tagName == `verse` {
				verseStr = p.findAttr(se, `number`)
				verseNum = p.findIntAttr(se, `number`)
			}
			if hasStyle[tagName] {
				usfmStyle = tagName + `.` + p.findAttr(se, `style`)
				if include(usfmStyle) { // This if fileters out the non-script code
					stack = stack.Push(usfmStyle)
				} else {
					err = decoder.Skip()
					if err != nil {
						status = log.Error(ctx, 500, err, "Error in USX parser")
						return records, titles, status
					}
				}
			}
		case xml.CharData:
			text := string(se)
			if len(strings.TrimSpace(text)) > 0 {
				// This is needed because scripts use {n} as a verse number.
				if strings.Contains(text, "{") || strings.Contains(text, "}") {
					text = strings.Replace(text, `{`, `(`, -1)
					text = strings.Replace(text, `}`, `)`, -1)
				}
				if usfmStyle == `para.h` {
					titles.heading = text
				} else if usfmStyle == `para.mt` || usfmStyle == `para.mt1` || usfmStyle == `para.mt2` || usfmStyle == `para.mt3` {
					titles.title = append(titles.title, text)
				} else {
					rec.ScriptTexts = append(rec.ScriptTexts, text)
				}
			}
		case xml.EndElement:
			if hasStyle[se.Name.Local] {
				stack, usfmStyle = stack.Pop()
			}
		}
		if chapterNum != rec.ChapterNum || verseNum != rec.VerseNum {
			if rec.BookId != `` && len(rec.ScriptTexts) > 0 {
				records = append(records, rec)
			}
			scriptNum++
			if chapterNum != rec.ChapterNum {
				scriptNum = 1
			}
			rec = db.Script{BookId: rec.BookId, ChapterNum: chapterNum, ScriptNum: strconv.Itoa(scriptNum),
				VerseNum: verseNum, VerseStr: verseStr, UsfmStyle: usfmStyle}
		}
	}
	if rec.BookId != `` && len(rec.ScriptTexts) > 0 {
		records = append(records, rec)
	}
	return records, titles, status
}

type titleDesc struct {
	heading string
	title   []string
}

func (p *USXParser) addChapterHeading(records []db.Script, titles titleDesc) []db.Script {
	var results = make([]db.Script, 0, len(records)+300)
	if len(records) == 0 {
		return results
	}
	var rec = records[0]
	rec.VerseStr = `0`
	rec.VerseNum = 0
	rec.UsfmStyle = `para.mt`
	rec.ScriptTexts = []string{strings.Join(titles.title, " ")}
	results = append(results, rec)
	var lastChapter = 1
	for _, rec = range records {
		if lastChapter != rec.ChapterNum {
			lastChapter = rec.ChapterNum
			var rec2 = rec
			rec2.VerseStr = `0`
			rec2.VerseNum = 0
			rec2.UsfmStyle = `para.h`
			rec2.ScriptTexts = []string{titles.heading + " " + strconv.Itoa(rec.ChapterNum)}
			results = append(results, rec2)
		}
		results = append(results, rec)
	}
	return results
}

func (p *USXParser) correctScriptNum(records []db.Script) []db.Script {
	var results = make([]db.Script, 0, len(records))
	var scriptNum = 0
	var lastChapter = 0
	for _, rec := range records {
		if rec.ChapterNum != lastChapter {
			lastChapter = rec.ChapterNum
			scriptNum = 0
		}
		scriptNum += 1
		rec.ScriptNum = strconv.Itoa(scriptNum)
		results = append(results, rec)
	}
	return results
}

func (p *USXParser) findAttr(se xml.StartElement, name string) string {
	for _, attr := range se.Attr {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ``
}

func (p *USXParser) findIntAttr(se xml.StartElement, name string) int {
	val := p.findAttr(se, name)
	if val == `` {
		return 0
	} else {
		return safe.SafeVerseNum(val)
	}
}

func (s Stack) Push(v string) Stack {
	return append(s, v)
}

func (s Stack) Pop() (Stack, string) {
	l := len(s)
	if l < 1 {
		panic("You tried to pop an empty stack.")
	}
	return s[:l-1], s[l-1]
}
