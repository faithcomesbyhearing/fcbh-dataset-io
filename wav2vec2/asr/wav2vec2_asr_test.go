package asr

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	"testing"
)

func TestWav2Vec2ASR_ProcessFiles(t *testing.T) {
	ctx := context.Background()
	//conn := db.NewDBAdapter(ctx, ":memory:")
	user := request.GetTestUser()
	conn, status := db.NewerDBAdapter(ctx, false, user, "N2KEUWB4")
	asr := NewWav2Vec2ASR(ctx, conn, "keu", "")
	var files []input.InputFile
	var file input.InputFile
	file.BookId = "LUK"
	file.Chapter = 1
	file.MediaId = "N2KEUWB4N2DA"
	files = append(files, file)
	status = asr.ProcessFiles(files)
	if status != nil {
		t.Fatal(status)
	}
}
