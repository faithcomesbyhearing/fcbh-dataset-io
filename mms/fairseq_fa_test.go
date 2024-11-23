package mms

import (
	"context"
	"dataset/db"
	"dataset/fetch"
	"dataset/input"
	"os"
	"path/filepath"
	"testing"
)

func TestFairseqFA_ProcessFiles(t *testing.T) {
	ctx := context.Background()
	//user, _ := fetch.GetTestUser()
	//conn, status := db.NewerDBAdapter(ctx, false, user.Username, "PlainTextEditScript_ENGWEB")
	dbPath := filepath.Join(os.Getenv("GOPROJ"), "dataset", "mms", "fairseq_fa", "mrk.sqlite")
	conn := db.NewDBAdapter(ctx, dbPath)
	count, status := conn.SelectScalarInt("select count(*) from scripts")
	if status.IsErr {
		t.Fatal(status)
	}
	if count == 0 {
		t.Fatal(status)
	}
	fa := NewForcedAlign(ctx, conn, "eng", "")
	var files []input.InputFile
	var file input.InputFile
	file.BookId = "MRK"
	file.Chapter = 1
	file.MediaId = "ENGWEBN2DA"
	file.Directory = os.Getenv("FCBH_DATASET_FILES") + "/ENGWEB/ENGWEBN2DA-mp3-64/"
	file.Filename = "B02___01_Mark________ENGWEBN2DA.mp3"
	files = append(files, file)
	status = fa.ProcessFiles(files)
	if status.IsErr {
		t.Fatal(status)
	}
}

func TestFairseqFA_processPyOutput(t *testing.T) {
	ctx := context.Background()
	user, _ := fetch.GetTestUser()
	conn, status := db.NewerDBAdapter(ctx, false, user.Username, "PlainTextEditScript_ENGWEB")
	fa := NewForcedAlign(ctx, conn, "eng", "")
	var file input.InputFile
	file.BookId = "MRK"
	file.Chapter = 1
	file.MediaId = "ENGWEBN2DA"
	file.Directory = os.Getenv("FCBH_DATASET_FILES") + "/ENGWEB/ENGWEBN2DA-mp3-64/"
	file.Filename = "B02___01_Mark________ENGWEBN2DA.mp3"
	outputFile := filepath.Join(os.Getenv("HOME"), "Desktop/manifest.json")
	references := []int64{1100, 1101, 1102, 1103, 1104, 1105, 1106, 1107, 1108, 1109, 1110, 1111,
		1112, 1113, 1114, 1115, 1116, 1117, 1118, 1119, 1120, 1121, 1122, 1123, 1124, 1125, 1126,
		1127, 1128, 1129, 1130, 1131, 1132, 1133, 1134, 1135, 1136, 1137, 1138, 1139, 1140, 1141,
		1142, 1143, 1144, 1145}
	status = fa.processPyOutput(file, outputFile, references)
	if status.IsErr {
		t.Fatal(status)
	}
}