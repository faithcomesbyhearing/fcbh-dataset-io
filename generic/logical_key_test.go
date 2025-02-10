package generic

import (
	"testing"
)

func TestLineRef_Parse(t *testing.T) {
	var a VerseRef
	a = NewVerseRef("NUM 22:12")
	if a.BookId != "NUM" {
		t.Error("BookId should be NUM")
	}
	if a.ChapterNum != 22 {
		t.Error("ChapterNum should be 22")
	}
	if a.VerseStr != "12" {
		t.Error("VerseStr should be 12")
	}
}

func TestLineRef_Compose(t *testing.T) {
	a := VerseRef{BookId: "NUM", ChapterNum: 22, VerseStr: "12"}
	b := a.UniqueKey()
	if b != "NUM 22:12" {
		t.Error("BookId should be NUM 22:12")
	}
}
