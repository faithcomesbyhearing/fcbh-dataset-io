package input

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"testing"
)

func TestPlainText1(t *testing.T) {
	ctx := context.Background()
	bibleId := `ENGWEB`
	fsType := request.TextPlain
	otFileset := `ENGWEBO_ET`
	ntFileset := `ENGWEBN_ET`
	//testament := request.Testament{NT: true, OTBooks: []string{`JOB`, `PSA`, `PRO`, `SNG`}}
	files, status := DBPDirectory(ctx, bibleId, fsType, otFileset, ntFileset)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 2 {
		t.Error(`There should be 2 files`)
	}
	if files[0].Filename != `ENGWEBO_ET.json` {
		t.Error(`First file should be ENGWEBO_ET.json`, files[0].Filename)
	}
	if files[0].Directory == `` {
		t.Error(`Directory should not be empty`)
	}
	//fmt.Println(files)
}

func TestPlainText2(t *testing.T) {
	ctx := context.Background()
	bibleId := `ENGWEB`
	fsType := request.TextPlain
	otFileset := ``
	ntFileset := `ENGWEBN_ET`
	//testament := request.Testament{NT: true}
	files, status := DBPDirectory(ctx, bibleId, fsType, otFileset, ntFileset)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 1 {
		t.Error(`There should be 1 file1`)
	}
	if files[0].Filename != `ENGWEBN_ET.json` {
		t.Error(`First file should be ENGWEBN_ET.json`, files[0].Filename)
	}
	if files[0].Directory == `` {
		t.Error(`Directory should not be empty`)
	}
	fmt.Println(files)
}

func TestUSXText1(t *testing.T) {
	ctx := context.Background()
	bibleId := `ENGWEB`
	fsType := request.TextUSXEdit
	otFileset := `ENGWEBO_ET-usx`
	ntFileset := `ENGWEBN_ET-usx`
	testament := request.Testament{NT: true, OTBooks: []string{`JOB`, `PSA`, `PRO`, `SNG`}}
	testament.BuildBookMaps()
	files, status := DBPDirectory(ctx, bibleId, fsType, otFileset, ntFileset)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 31 {
		t.Error(`There should be 31 files`, len(files))
	}
	if files[0].Filename != `018JOB.usx` {
		t.Error(`First file should be 018JOB.usx`, files[0].Filename)
	}
	if files[0].Directory == `` {
		t.Error(`Directory should not be empty`)
	}
	if files[0].BookId != `JOB` {
		t.Error(`First book id should be JOB`)
	}
	if files[4].BookId != `MAT` {
		t.Error(`Fifth book id should be MAT`)
	}
	//fmt.Println(files)
}

func TestUSXText2(t *testing.T) {
	ctx := context.Background()
	bibleId := `ENGWEB`
	fsType := request.TextUSXEdit
	otFileset := `ENGWEBO_ET-usx`
	ntFileset := ``
	testament := request.Testament{OTBooks: []string{`JOB`, `PSA`, `PRO`, `SNG`}}
	testament.BuildBookMaps()
	files, status := DBPDirectory(ctx, bibleId, fsType, otFileset, ntFileset)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 4 {
		t.Error(`There should be 4 files`, len(files))
	}
	if files[0].Filename != `018JOB.usx` {
		t.Error(`First file should be 018JOB.usx`, files[0].Filename)
	}
	if files[0].Directory == `` {
		t.Error(`Directory should not be empty`)
	}
	if files[0].BookId != `JOB` {
		t.Error(`First book id should be JOB`)
	}
	if files[3].BookId != `SNG` {
		t.Error(`Fifth book id should be SNG`)
	}
	//fmt.Println(files)
}

func TestAudio1(t *testing.T) {
	ctx := context.Background()
	bibleId := `ENGWEB`
	fsType := request.Audio //`audio`
	otFileset := ``
	ntFileset := `ENGWEBN2DA-mp3-64`
	testament := request.Testament{NTBooks: []string{`ROM`, `EPH`, `COL`, `HEB`}}
	testament.BuildBookMaps()
	files, status := DBPDirectory(ctx, bibleId, fsType, otFileset, ntFileset)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 39 {
		t.Error(`There should be 39 files`, len(files))
	}
	if files[0].Filename != `B06___01_Romans______ENGWEBN2DA.mp3` {
		t.Error(`First file should be B06___01_Romans______ENGWEBN2DA.mp3`, files[0].Filename)
	}
	if files[0].Directory == `` {
		t.Error(`Directory should not be empty`)
	}
	if files[0].BookId != `ROM` {
		t.Error(`First book id should be MAT`)
	}
	if files[4].Chapter != 5 {
		t.Error(`Fifth file should be chapter 5`)
	}
	//fmt.Println(files)
}

func TestAudio2(t *testing.T) {
	ctx := context.Background()
	bibleId := `ENGWEB`
	fsType := request.Audio
	otFileset := ``
	ntFileset := `ENGWEBN2DA-opus16`
	testament := request.Testament{NT: true, NTBooks: []string{`ROM`, `EPH`, `COL`, `HEB`}}
	testament.BuildBookMaps()
	files, status := DBPDirectory(ctx, bibleId, fsType, otFileset, ntFileset)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 260 {
		t.Error(`There should be 260 files`, len(files))
	}
	if files[0].Filename != `B01___01_Matthew_____ENGWEBN2DA.webm` {
		t.Error(`First file should be ENGWEBN2DA.webm`, files[0].Filename)
	}
	if files[0].Directory == `` {
		t.Error(`Directory should not be empty`)
	}
	if files[0].BookId != `MAT` {
		t.Error(`First book id should be MAT`)
	}
	if files[4].Chapter != 5 {
		t.Error(`Fifth file should be chapter 5`)
	}
	if files[4].BookId != `MAT` {
		t.Error(`Fifth file should be MAT`)
	}
	last := files[len(files)-1]
	if last.BookId != `REV` {
		t.Error(`The last bookId should be REV`)
	}
	if last.Chapter != 22 {
		t.Error(`The last chapter should be 22`)
	}
	//fmt.Println(files)
}

func TestIncorrectFileset(t *testing.T) {
	ctx := context.Background()
	bibleId := `ENGWEB`
	fsType := request.Audio
	otFileset := ``
	ntFileset := `ENGWEBN2DA-opus99`
	testament := request.Testament{NT: true, NTBooks: []string{`ROM`, `EPH`, `COL`, `HEB`}}
	testament.BuildBookMaps()
	files, status := DBPDirectory(ctx, bibleId, fsType, otFileset, ntFileset)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 0 {
		t.Error(`There should be 0 files`, len(files))
	}
	//fmt.Println(files)
}

func TestIncorrectBibleId(t *testing.T) {
	ctx := context.Background()
	bibleId := `ENGBEW22`
	fsType := request.Audio
	otFileset := ``
	ntFileset := `ENGWEBN2DA-mp3-64`
	testament := request.Testament{NT: true, NTBooks: []string{`ROM`, `EPH`, `COL`, `HEB`}}
	testament.BuildBookMaps()
	files, status := DBPDirectory(ctx, bibleId, fsType, otFileset, ntFileset)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 0 {
		t.Error(`There should be 0 files`, len(files))
	}
	fmt.Println(files)
}

func TestIncorrectBooks(t *testing.T) {
	ctx := context.Background()
	bibleId := `ENGWEB`
	fsType := request.Audio
	otFileset := ``
	ntFileset := `ENGWEBN2DA-mp3-64`
	testament := request.Testament{NTBooks: []string{`RO1`, `EP1`, `CO1`, `HE1`}}
	testament.BuildBookMaps()
	files, status := DBPDirectory(ctx, bibleId, fsType, otFileset, ntFileset)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 0 {
		t.Error(`There should be 0 files`, len(files))
	}
	//fmt.Println(files)
}

func TestParseV4AudioFilename(t *testing.T) {
	ctx := context.Background()
	var status *log.Status
	var file InputFile
	file.Filename = `ENGESVN2DA_B001_MAT_001.mp3`
	status = parseV4AudioFilename(ctx, &file)
	if status != nil {
		t.Error(status)
	}
	if file.MediaId != `ENGESVN2DA` {
		t.Error(`mediaId is incorrect`, file.MediaId)
	}
	if file.Testament != `NT` {
		t.Error(`Testament is incorrect`, file.Testament)
	}
	if file.BookId != `MAT` {
		t.Error(`BookId is incorrect`, file.BookId)
	}
	if file.BookSeq != `001` {
		t.Error(`BookSeq is incorrect`, file.BookSeq)
	}
	if file.Chapter != 1 {
		t.Error(`Chapter is incorrect`, file.Chapter)
	}
	if file.Verse != `` {
		t.Error(`Verse is incorrect`, file.Verse)
	}
	//fmt.Println("File", file)
	var file2 InputFile
	file2.Filename = `IRUNLCP1DA_B013_1TH_001_001-001_010.mp3`
	status = parseV4AudioFilename(ctx, &file2)
	if status != nil {
		t.Error(status)
	}
	if file2.Verse != `001` {
		t.Error(`Verse is incorrect`, file2.Verse)
	}
	if file2.ChapterEnd != 1 {
		t.Error(`ChapterEnd is incorrect`, file2.ChapterEnd)
	}
	if file2.VerseEnd != `010` {
		t.Error(`VerseEnd is incorrect`, file2.VerseEnd)
	}
	//fmt.Println("File2", file2)
}
