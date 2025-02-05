package read

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/cli_misc"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"github.com/xuri/excelize/v2"
	"os"
	"path/filepath"
	"testing"
)

func TestScriptReader(t *testing.T) {
	bibleId := `ATIWBT`
	database := bibleId + "_SCRIPT.db"
	db.DestroyDatabase(database)
	conn := db.NewDBAdapter(context.Background(), database)
	testament := request.Testament{OT: true, NT: true}
	script := NewScriptReader(conn, testament)
	filename := filepath.Join(os.Getenv(`FCBH_DATASET_FILES`), bibleId, bibleId+`N2ST.xlsx`)
	fmt.Println(`Filename:`, filename)
	status := script.Read(filename)
	if status != nil {
		t.Fatal(status)
	}
	conn.Close()
}

// TestScriptHeaders is to investigate the format of script columns for position of items
func TestScriptHeaders(t *testing.T) {
	ctx := context.Background()
	ts := cli_misc.NewTSBucket(ctx)
	list := ts.ListPrefix(cli_misc.TSBucketName, cli_misc.LatinN2)
	for count, item := range list {
		if count > 400 {
			break
		}
		//fmt.Println(item)
		objs := ts.ListObjects(cli_misc.TSBucketName, item+cli_misc.Script)
		key := objs[0]
		//fmt.Println(key)
		filename := filepath.Base(key)
		filePath := filepath.Join(os.Getenv(`FCBH_DATASET_TMP`), filename)
		//fmt.Println(filePath)
		ts.DownloadObject(cli_misc.TSBucketName, key, filePath)
		file, err := excelize.OpenFile(filePath)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		sheets := file.GetSheetList()
		rows, err := file.GetRows(sheets[0])
		if err != nil {
			t.Fatal(err)
		}
		for col, cell := range rows[0] {
			fmt.Print(col, `:`, cell, ` `)
		}
		fmt.Print("\n")
		_ = file.Close()
		_ = os.Remove(filePath)
	}
}
