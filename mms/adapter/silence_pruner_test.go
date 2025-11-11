package adapter

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"sort"
	"strconv"
	"testing"
)

func TestSilencePruner(t *testing.T) {
	ctx := context.Background()
	log.SetOutput("stderr")
	user := request.GetTestUser()
	conn, status := db.NewerDBAdapter(ctx, false, user, "N2MZJSIM")
	if status != nil {
		t.Fatal(status)
	}
	threshold := 400
	status = SilencePruner(ctx, threshold, conn)
	if status != nil {
		t.Fatal(status)
	}
	rows, err := conn.DB.Query("SELECT script_id FROM pruned_silence")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var count int
	var scriptId int64
	for rows.Next() {
		rows.Scan(&scriptId)
		//fmt.Println(scriptId)
		count++
	}
	err = rows.Err()
	if err != nil {
		t.Fatal(err)
	}
	if count != threshold {
		t.Error("Count should be ", threshold)
	}
}

func TestFindSilence(t *testing.T) {
	ctx := context.Background()
	log.SetOutput("stderr")
	user := request.GetTestUser()
	conn, status := db.NewerDBAdapter(ctx, false, user, "N2MZJSIM")
	if status != nil {
		t.Fatal(status)
	}
	silences := findSilence(ctx, 400, conn)
	sort.Slice(silences, func(i, j int) bool {
		return silences[i].WordId < silences[j].WordId
	})
	numMissing := findMissingVerses(silences)
	fmt.Println("numMissing", numMissing)
	//dumpTimestamps(silences)
}

func dumpTimestamps(silences []SilenceRec) {
	for i := 0; i < 5000; i++ {
		s := silences[i]
		fmt.Println(s.Ref(), s.BeginT(), s.EndT(), s.EndScript(), s.Silen(), s.LastWord)
	}
}

type ref struct {
	bookId     string
	chapterNum int
	verseStr   string
}

func findMissingVerses(silences []SilenceRec) int {
	var refMap = make(map[ref]bool)
	refMap[ref{bookId: "MAT", chapterNum: 5, verseStr: "36"}] = false
	refMap[ref{bookId: "MAT", chapterNum: 8, verseStr: "9"}] = false
	refMap[ref{bookId: "MAT", chapterNum: 12, verseStr: "33"}] = false
	refMap[ref{bookId: "MAT", chapterNum: 12, verseStr: "37"}] = false
	refMap[ref{bookId: "MAT", chapterNum: 17, verseStr: "22"}] = false
	refMap[ref{bookId: "MAT", chapterNum: 22, verseStr: "32"}] = false
	refMap[ref{bookId: "MAT", chapterNum: 25, verseStr: "38"}] = false
	refMap[ref{bookId: "MAT", chapterNum: 26, verseStr: "6"}] = false
	refMap[ref{bookId: "MRK", chapterNum: 4, verseStr: "5"}] = false
	refMap[ref{bookId: "MRK", chapterNum: 4, verseStr: "24"}] = false
	refMap[ref{bookId: "MRK", chapterNum: 9, verseStr: "39"}] = false
	refMap[ref{bookId: "MRK", chapterNum: 10, verseStr: "38"}] = false
	refMap[ref{bookId: "LUK", chapterNum: 4, verseStr: "19"}] = false
	refMap[ref{bookId: "LUK", chapterNum: 11, verseStr: "42"}] = false
	refMap[ref{bookId: "LUK", chapterNum: 20, verseStr: "38"}] = false
	refMap[ref{bookId: "LUK", chapterNum: 23, verseStr: "29"}] = false
	refMap[ref{bookId: "JHN", chapterNum: 3, verseStr: "21"}] = false
	refMap[ref{bookId: "JHN", chapterNum: 5, verseStr: "22"}] = false
	refMap[ref{bookId: "JHN", chapterNum: 19, verseStr: "6"}] = false
	refMap[ref{bookId: "JHN", chapterNum: 21, verseStr: "19"}] = false
	refMap[ref{bookId: "ACT", chapterNum: 8, verseStr: "39"}] = false
	refMap[ref{bookId: "ACT", chapterNum: 12, verseStr: "10"}] = false
	refMap[ref{bookId: "EPH", chapterNum: 3, verseStr: "6"}] = false
	for _, s := range silences {
		key := ref{bookId: s.BookId, chapterNum: s.ChapterNum, verseStr: s.VerseStr}
		_, found := refMap[key]
		if found {
			refMap[key] = true
		}
	}
	var count = 0
	for key, found := range refMap {
		if !found {
			fmt.Println("Not found", key)
			count++
		}
	}
	return count
}

func (s *SilenceRec) Ref() string {
	return s.BookId + " " + strconv.Itoa(s.ChapterNum) + ":" + s.VerseStr + "." + strconv.Itoa(s.WordSeq)
}
func (s *SilenceRec) BeginT() string {
	return strconv.FormatFloat(s.BeginTS, 'f', 2, 64)
}
func (s *SilenceRec) EndT() string {
	return strconv.FormatFloat(s.EndTS, 'f', 2, 64)
}
func (s *SilenceRec) EndScript() string {
	return strconv.FormatFloat(s.ScriptEndTS, 'f', 2, 64)
}
func (s *SilenceRec) Dur() string {
	return strconv.FormatFloat(s.Duration, 'f', 2, 64)
}
func (s *SilenceRec) Silen() string {
	return strconv.FormatFloat(s.Silence, 'f', 2, 64)
}
