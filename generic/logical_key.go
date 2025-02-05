package generic

import (
	"strconv"
	"strings"
)

// The LogicalKey interface and LineRef type are an experimental idea
// for how to abstract out the details for bible references.
// So, that any module that was not specifically about the Bible could use
// The logical key.  This approach would allow modules that are not specifically
// for the Bible to use a Bible reference and be certain it has a consistent meaning.

type LogicalKey interface {
	IsLogicalKey()
	UniqueKey() string
	Description() string
}

type VerseRef struct {
	BookId     string
	ChapterNum int
	VerseStr   string
	ChapterEnd int
	VerseEnd   string
}

func (r VerseRef) IsLogicalKey() {}

func (r VerseRef) UniqueKey() string {
	var result string
	if r.ChapterNum == 0 {
		result = r.BookId
	} else if r.VerseStr == `` {
		result = r.BookId + ` ` + strconv.Itoa(r.ChapterNum)
	} else {
		result = r.BookId + ` ` + strconv.Itoa(r.ChapterNum) + `:` + r.VerseStr
	}
	return result
}

func (r VerseRef) Description() string {
	var result string
	if r.ChapterNum == 0 {
		result = r.BookId
	} else if r.VerseStr == `` {
		result = r.BookId + ` ` + strconv.Itoa(r.ChapterNum)
	} else if (r.ChapterEnd == 0 || r.ChapterEnd == r.ChapterNum) &&
		(r.VerseEnd == `` || r.VerseEnd == r.VerseStr) {
		result = r.BookId + ` ` + strconv.Itoa(r.ChapterNum) + `:` + r.VerseStr
	} else if r.ChapterEnd == 0 || r.ChapterEnd == r.ChapterNum {
		result = r.BookId + ` ` + strconv.Itoa(r.ChapterNum) + `:` + r.VerseStr + `-` + r.VerseEnd
	} else {
		result = r.BookId + ` ` + strconv.Itoa(r.ChapterNum) + `:` + r.VerseStr + `-` + strconv.Itoa(r.ChapterEnd) + `:` + r.VerseEnd
	}
	return result
}

func NewLineRef(key string) VerseRef {
	var r VerseRef
	parts := strings.Split(key, ` `)
	r.BookId = parts[0]
	if len(parts) > 1 {
		parts = strings.Split(parts[1], `-`)
		pieces := strings.Split(parts[0], `:`)
		r.ChapterNum, _ = strconv.Atoi(pieces[0])
		if len(pieces) > 1 {
			r.VerseStr = pieces[1]
		}
		if len(parts) > 1 {
			pieces = strings.Split(parts[0], `:`)
			if len(pieces) == 1 {
				r.VerseEnd = pieces[0]
			} else {
				r.ChapterEnd, _ = strconv.Atoi(pieces[0])
				r.VerseEnd = pieces[1]
			}
		}
	}
	return r
}

////////////////// Example of other Use of Logical Key ///////////////

type Publish struct {
	BookId     string
	ChapterNum int
	Page       int
	Paragraph  int
	Sentence   int
	sequence   int
}

func (p Publish) IsLogicalKey()       {}
func (p Publish) UniqueKey() string   { return "" }
func (p Publish) Description() string { return "" }
func NewPublish(key string) Publish   { return Publish{} }
