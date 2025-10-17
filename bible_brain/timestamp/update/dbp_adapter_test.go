package update

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/safe"
)

func TestNewDBPAdapter(t *testing.T) {
	conn := getDBPConnection(t)
	conn.Close()
}

func getDBPConnection(t *testing.T) DBPAdapter {
	ctx := context.Background()
	conn, status := NewDBPAdapter(ctx)
	if status != nil {
		t.Fatal(status)
	}
	return conn
}

func TestSelectHash(t *testing.T) {
	conn := getDBPConnection(t)
	defer conn.Close()
	hashId, status := conn.SelectHashId("ENGKJVN2DA")
	if status != nil {
		t.Fatal(status)
	}
	fmt.Println("hashId", hashId)
	if hashId != "84121069c3cc" {
		t.Fatal("hashId should be 84121069c3cc")
	}
}

func TestSelectFileId(t *testing.T) {
	conn := getDBPConnection(t)
	defer conn.Close()
	hashId, status := conn.SelectHashId("ENGKJVN2DA")
	if status != nil {
		t.Fatal(status)
	}
	fileId, audioFile, status := conn.SelectFileId(hashId, "MAT", 1)
	if status != nil {
		t.Fatal(status)
	}
	fmt.Println("audioFile", audioFile, "fileId", fileId)
	if fileId != 788486 {
		t.Fatal("fileId should be 614190, but is", fileId)
	}
}

func TestSelectTimestamps(t *testing.T) {
	conn := getDBPConnection(t)
	defer conn.Close()
	hashId, status := conn.SelectHashId("ENGKJVN2DA")
	if status != nil {
		t.Fatal(status)
	}
	fileId, _, status := conn.SelectFileId(hashId, "MAT", 1)
	if status != nil {
		t.Fatal(status)
	}
	timestamps, status := conn.SelectTimestamps(fileId)
	if status != nil {
		t.Fatal(status)
	}
	for _, ts := range timestamps {
		fmt.Println(ts)
	}
	if len(timestamps) != 26 {
		t.Fatal("timestamps length should be 26, but is ", len(timestamps))
	}
	if timestamps[25].VerseStr != "25" {
		t.Fatal("VerseStr should be 25")
	}
}

func TestUpdateFilesetTimingEstTag(t *testing.T) {
	conn := getDBPConnection(t)
	defer conn.Close()
	hashId, status := conn.SelectHashId("ENGKJVN2DA")
	if status != nil {
		t.Fatal(status)
	}
	rowCount, status := conn.UpdateFilesetTimingEstTag(hashId, mmsAlignTimingEstErr)
	if status != nil {
		t.Fatal(status)
	}
	if rowCount != 1 {
	}
}

func fauxTimesheetData(timestamps []Timestamp) []Timestamp {
	var priorTS = 0.0
	var lastVerse string
	var lastSeq int
	for i := range timestamps {
		timestamps[i].TimestampId = 0
		timestamps[i].BeginTS = priorTS
		priorTS = float64(i) * 1.2
		timestamps[i].EndTS = priorTS
		lastVerse = timestamps[i].VerseStr
		lastSeq = timestamps[i].VerseSeq
	}
	verseNum := strconv.Itoa(safe.SafeVerseNum(lastVerse) + 1)
	var ts Timestamp
	ts.VerseStr = verseNum
	ts.VerseSeq = lastSeq + 1
	ts.BeginTS = priorTS
	ts.EndTS = (float64(len(timestamps)) + 1.0) * 1.2
	timestamps = append(timestamps, ts)
	return timestamps
}
