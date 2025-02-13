package update

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/safe"
	"strconv"

	//"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"testing"
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

func TestUpdateTimestamps(t *testing.T) {
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
	dbpTimestamps, status := conn.SelectTimestamps(fileId)
	if status != nil {
		t.Fatal(status)
	}
	timestamps := FauxTimesheetData(dbpTimestamps)
	// Remove some DBP Records
	var dbp2Timestamps []Timestamp
	for i := 0; i < len(dbpTimestamps); i += 2 {
		dbp2Timestamps = append(dbp2Timestamps, dbpTimestamps[i])
	}
	timestamps = MergeTimestamps(timestamps, dbp2Timestamps)
	rowCount, status := conn.UpdateTimestamps(timestamps)
	if status != nil {
		t.Fatal(status)
	}
	if rowCount != 12 {
		t.Fatal("rowCount should be 12, but is", rowCount)
	}
}

/*
func TestInsertTimestamps(t *testing.T) {
	bookId := "MAT"
	chapter := 1
	ctx := context.Background()
	var datasetConn UpdateTimestamps
	var status *log.Status
	datasetConn.conn, status = db.NewerDBAdapter(ctx, true, user, project)
	if status != nil {
		t.Fatal(status)
	}
	var timestamps []Timestamp
	timestamps, status = datasetConn.SelectTimestamps(bookId, chapter)
	// select timestamps from
	conn := getDBPConnection(t)
	defer conn.Close()

	timestamps, status := conn.SelectTimestamps("ENGWEBN2DA", "MAT", 1)
	if status != nil {
		t.Fatal(status)
	}
	for i := range timestamps {
		timestamps[i].BeginTS = timestamps[i].BeginTS * 2.0
		timestamps[i].EndTS = timestamps[i].EndTS * 2.0
	}
	times, _, status := conn.InsertTimestamps(timestamps)
	if status != nil {
		t.Fatal(status)
	}
	if len(times) != len(timestamps) {
		t.Error("len times and timestamps not equal")
	}
	for i := range times {
		if times[i].BeginTS != timestamps[i].BeginTS*2.0 {
			t.Error("Update was to double times")
		}
	}
}

func TestUpdateFilesetTimingEstTag(t *testing.T) {
	conn := getDBPConnection(t)
	defer conn.Close()
	hashId, status := conn.SelectHashId("ENGWEBN2DA")
	if status != nil {
		t.Fatal(status)
	}
	status = conn.UpdateFilesetTimingEstTag(hashId, "Bob") //mmsAlignTimingEstErr)
	if status != nil {
		t.Fatal(status)
	}
}
*/
//InsertBandwidth
//InsertSegments

//func TestDeleteTimestamps()

func FauxTimesheetData(timestamps []Timestamp) []Timestamp {
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
