package mms

import (
	"context"
	"dataset/db"
	"dataset/input"
	"os"
	"testing"
)

func TestMMSASR_ProcessFiles(t *testing.T) {
	ctx := context.Background()
	conn := db.NewDBAdapter(ctx, ":memory:")
	asr := NewMMSASR(ctx, conn, "eng", "")
	var files []input.InputFile
	var file input.InputFile
	file.BookId = "MRK"
	file.Chapter = 1
	file.MediaId = "ENGWEBN2DA"
	file.Directory = os.Getenv("FCBH_DATASET_FILES") + "/ENGWEB/ENGWEBN2DA-mp3-64/"
	file.Filename = "B02___01_Mark________ENGWEBN2DA.mp3"
	//file.MediaId = "ENGESVN1DA"
	//file.Directory = os.Getenv("FCBH_DATASET_FILES") + "/ENGESV/ENGESVN1DA/"
	//file.Filename = "B02___01_Mark________ENGESVN1DA.mp3"
	files = append(files, file)
	status := asr.ProcessFiles(files)
	if status.IsErr {
		t.Fatal(status)
	}
}
