package input

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"testing"
)

func TestUtility_validateBookId(t *testing.T) {
	ctx := context.Background()
	bookId, status := validateBookId(ctx, "TTL")
	if status != nil {
		t.Error(status)
	}
	if bookId != "TIT" {
		t.Error(bookId, "should have been revised to TIT")
	}
}

func TestUtility_parseFilenames(t *testing.T) {
	ctx := context.Background()
	test1 := InputFile{MediaType: request.TextUSXEdit, Directory: "/ABC/DEF", Filename: "001GEN.usx"}
	status := parseFilenames(ctx, &test1)
	if status != nil {
		t.Error(status)
	}
	if test1.MediaId != "DEF" {
		t.Error("Media ID should be DEF")
	}
	if test1.BookId != "GEN" {
		t.Error("Book ID should be GEN")
	}
	if test1.BookSeq != "001" {
		t.Error("Book Seq should be 001")
	}
	test2 := InputFile{MediaType: request.TextUSXEdit, Directory: "/ABC/DEF", Filename: "GEN.usx"}
	status = parseFilenames(ctx, &test2)
	if status != nil {
		t.Error(status)
	}
	if test2.BookId != "GEN" {
		t.Error("Book ID should be GEN")
	}
	if test2.BookSeq != "1" {
		t.Error("Book Seq should be 1")
	}
	test3 := InputFile{MediaType: request.TextUSXEdit, Directory: "/ABC/DEF", Filename: "1GEN.usx"}
	status = parseFilenames(ctx, &test3)
	if status == nil {
		t.Error(status)
	}
}
