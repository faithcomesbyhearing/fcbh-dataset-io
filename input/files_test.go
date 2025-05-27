package input

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"os"
	"path/filepath"
	"testing"
)

func TestPlainTextFileInput(t *testing.T) {
	ctx := context.Background()
	bibleId := `ENGWEB`
	testament := request.Testament{OTBooks: []string{`MAL`, `JON`}, NTBooks: []string{`TIT`, `REV`}}
	testament.BuildBookMaps()
	search := filepath.Join(os.Getenv(`FCBH_DATASET_FILES`), bibleId, `*_ET.json`)
	files, status := FileInput(ctx, search)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 2 {
		t.Error(`Test should have found 2 files`, len(files))
	}
}

func TestUSXTextFileInput(t *testing.T) {
	ctx := context.Background()
	bibleId := `ENGWEB`
	filesetId := `ENGWEBO_ET-usx`
	testament := request.Testament{OTBooks: []string{`MAL`, `JON`, `GEN`}, NTBooks: []string{`TIT`, `REV`}}
	testament.BuildBookMaps()
	search := filepath.Join(os.Getenv(`FCBH_DATASET_FILES`), bibleId, filesetId, `*.usx`)
	files, status := FileInput(ctx, search)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 3 {
		t.Error(`Test should have found 3 files`, len(files))
	}
}

func TestAudioFileInput(t *testing.T) {
	ctx := context.Background()
	bibleId := `ENGWEB`
	filesetId := `ENGWEBN2DA-mp3-64`
	testament := request.Testament{OTBooks: []string{`MAL`, `JON`, `GEN`}, NTBooks: []string{`TIT`, `REV`}}
	testament.BuildBookMaps()
	search := filepath.Join(os.Getenv(`FCBH_DATASET_FILES`), bibleId, filesetId, `*.mp3`)
	files, status := FileInput(ctx, search)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 25 {
		t.Error(`Test should have found 25 files`, len(files))
	}
	//fmt.Println(files)
}
