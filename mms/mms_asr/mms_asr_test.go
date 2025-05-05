package mms_asr

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	"os"
	"testing"
)

func TestMMSASR_ProcessFiles(t *testing.T) {
	ctx := context.Background()
	//conn := db.NewDBAdapter(ctx, ":memory:")
	user := request.GetTestUser()
	conn, status := db.NewerDBAdapter(ctx, false, user, "PlainTextEditScript_ENGWEB")
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
	status = asr.ProcessFiles(files)
	if status != nil {
		t.Fatal(status)
	}
}
