package diff

import (
	"database/sql"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/generic"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type Pair struct {
	Ref       generic.VerseRef      `json:"ref"`
	ScriptNum string                `json:"script_num"`
	BeginTS   float64               `json:"begin_ts"`
	EndTS     float64               `json:"end_ts"`
	Base      PairText              `json:"base"`
	Comp      PairText              `json:"comp"`
	Diffs     []diffmatchpatch.Diff `json:"diffs"`
	HTML      string                `json:"html""`
}

type PairText struct {
	ScriptId int    `json:"script_id"`
	Text     string `json:"text"`
	Uroman   string `json:"uroman"`
}

func NewPair(base *Verse, comp *Verse) Pair {
	var p Pair
	if base != nil {
		p.Ref.BookId = base.bookId
		p.Ref.ChapterNum = base.chapter
		p.Ref.ChapterEnd = base.chapterEnd
		p.Ref.VerseStr = base.verse
		p.Ref.VerseEnd = base.verseEnd
		p.ScriptNum = base.scriptNum
		p.BeginTS = base.beginTS
		p.EndTS = base.endTS
		p.Base.ScriptId = base.scriptId
		p.Base.Text = base.text
		p.Base.Uroman = base.uRoman
	}
	if comp != nil {
		if p.Ref.BookId == "" {
			p.Ref.BookId = comp.bookId
			p.Ref.ChapterNum = comp.chapter
			p.Ref.ChapterEnd = comp.chapterEnd
			p.Ref.VerseStr = comp.verse
			p.Ref.VerseEnd = comp.verseEnd
			p.ScriptNum = comp.scriptNum
			if comp.beginTS != 0.0 {
				p.BeginTS = comp.beginTS
			}
			if comp.endTS != 0.0 {
				p.EndTS = comp.endTS
			}
		}
		p.Comp.ScriptId = comp.scriptId
		p.Comp.Text = comp.text
		p.Comp.Uroman = comp.uRoman
	}
	return p
}

func (p *Pair) ScriptId() int {
	if p.Base.ScriptId != 0 {
		return p.Base.ScriptId
	} else {
		return p.Comp.ScriptId
	}
}

func (p *Pair) Text(isLatin sql.NullBool) (string, string) {
	if isLatin.Bool {
		return p.Base.Text, p.Comp.Text
	} else {
		return p.Base.Uroman, p.Comp.Uroman
	}
}

func (p *Pair) Inserts() int {
	var inserts int
	for _, diff := range p.Diffs {
		lenText := len(diff.Text)
		if diff.Type == diffmatchpatch.DiffInsert {
			inserts += lenText
		}
	}
	return inserts
}

func (p *Pair) Deletes() int {
	var deletes int
	for _, diff := range p.Diffs {
		lenText := len(diff.Text)
		if diff.Type == diffmatchpatch.DiffDelete {
			deletes += lenText
		}
	}
	return deletes
}

func (p *Pair) LargestLength() int {
	var result int
	var length int
	for _, diff := range p.Diffs {
		if diff.Type != diffmatchpatch.DiffEqual {
			length += len(diff.Text)
		} else {
			if length > result {
				result = length
			}
			length = 0
		}
	}
	if length > result {
		result = length
	}
	return result
}

func (p *Pair) ErrorPct(inserts int, deletes int) float64 {
	avgLen := float64(len(p.Base.Uroman)+len(p.Comp.Uroman)) / 2.0
	return float64((inserts+deletes)*100) / avgLen
}
