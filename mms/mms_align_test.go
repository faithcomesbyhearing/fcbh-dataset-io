package mms

import (
	"context"
	"dataset/db"
	"dataset/decode_yaml/request"
	"dataset/input"
	log "dataset/logger"
	"dataset/utility"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// These tests are dependent upon test 02_plain_text_edit_script_test.go
// which creates the database: /Users/gary/FCBH2024/GaryNTest/01c_usx_text_edit_ENGWEB.db
// It is best to rerun test 02 in order to have a clean database

func TestMMSFA_ProcessFiles(t *testing.T) {
	ctx := context.Background()
	user := request.GetTestUser()
	conn, status := db.NewerDBAdapter(ctx, false, user, "01c_usx_text_edit_ENGWEB")
	if status != nil {
		t.Fatal(status)
	}
	fa := NewMMSAlign(ctx, conn, "eng", "")
	var files []input.InputFile
	var file input.InputFile
	file.BookId = "MAT"
	file.Chapter = 22
	file.MediaId = "ENGWEBN2DA"
	file.Directory = os.Getenv("FCBH_DATASET_FILES") + "/ENGWEB/ENGWEBN2DA-mp3-64/"
	//file.Filename = "B02___22_Mark________ENGWEBN2DA.mp3"
	file.Filename = "B01___22_Matthew_____ENGWEBN2DA.mp3"
	files = append(files, file)
	status = fa.ProcessFiles(files)
	if status != nil {
		t.Fatal(status)
	}
}

func TestMMSFA_prepareText(t *testing.T) {
	ctx := context.Background()
	database := "../match/N2HOYWFW.db"
	conn := db.NewDBAdapter(ctx, database)
	fa := NewMMSAlign(ctx, conn, "eng", "")
	var status *log.Status
	uromanPath := filepath.Join(os.Getenv("GOPROJ"), "dataset", "mms", "uroman_stdio.py")
	fa.uroman, status = utility.NewStdioExec(ctx, os.Getenv(`FCBH_MMS_FA_PYTHON`), uromanPath)
	if status != nil {
		t.Fatal(status)
	}
	for _, bookId := range db.BookNT {
		lastChap := db.BookChapterMap[bookId]
		for chap := 1; chap <= lastChap; chap++ {
			textList, refList, status := fa.prepareText("eng", bookId, chap)
			if status != nil {
				t.Fatal(status)
			}
			fmt.Println(bookId, chap, len(textList), len(refList))
		}
	}
}

func TestMMSFA_processPyOutput(t *testing.T) {
	ctx := context.Background()
	user := request.GetTestUser()
	conn, status := db.NewerDBAdapter(ctx, false, user, "01c_usx_text_edit_ENGWEB")
	if status != nil {
		t.Fatal(status)
	}
	fa := NewMMSAlign(ctx, conn, "eng", "")
	uromanPath := filepath.Join(os.Getenv("GOPROJ"), "dataset", "mms", "uroman_stdio.py")
	fa.uroman, status = utility.NewStdioExec(ctx, os.Getenv(`FCBH_MMS_FA_PYTHON`), uromanPath)
	if status != nil {
		t.Fatal(status)
	}
	var file input.InputFile
	file.BookId = "TIT"
	file.Chapter = 1
	file.MediaId = "ENGWEBN2DA"
	file.Directory = os.Getenv("FCBH_DATASET_FILES") + "/ENGWEB/ENGWEBN2DA-mp3-64/"
	//file.Filename = "B02___01_Mark________ENGWEBN2DA.mp3"
	//file.Filename = "B01___22_Matthew_____ENGWEBN2DA.mp3" // I think it doesn't matter
	file.Filename = "TestFile.mp3"
	var wordList []Word
	_, wordList, status = fa.prepareText("eng", file.BookId, file.Chapter)
	if status != nil {
		t.Fatal(status)
	}
	bytes, err := os.ReadFile("engweb_fa_out.json")
	if err != nil {
		t.Fatal(err)
	}
	status = fa.processPyOutput(file, wordList, string(bytes))
	if status != nil {
		t.Fatal(status)
	}
	scriptRows, status := conn.SelectScalarInt("select count(*) from scripts where script_end_ts != 0.0")
	if status != nil {
		t.Fatal(status)
	}
	if scriptRows != 46 {
		t.Error("scriptRows is", scriptRows, "it should be 46")
	}
	wordRows, status := conn.SelectScalarInt("select count(*) from words where fa_score != 0.0")
	if status != nil {
		t.Fatal(status)
	}
	if wordRows != 882 {
		t.Error("wordRows is", wordRows, "it should be 882")
	}
}
