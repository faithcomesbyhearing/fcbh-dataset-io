package tests

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/controller"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"strings"
	"testing"
)

const scriptTextScript = `is_new: yes
dataset_name: 01d_script_text_{bibleId}
bible_id: {bibleId}
username: GaryNTest
output:
  sqlite: yes
text_data:
  file: /Users/gary/FCBH2024/download/ATIWBT/ATIWBTN2ST.xlsx
`

func TestScriptTextDirect(t *testing.T) {
	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 9747})
	testName := strings.Replace(scriptTextScript, "{bibleId}", "ENGWEB", -1)
	DirectSqlTest(testName, tests, t)
}

func TestScriptTextScriptAPI(t *testing.T) {
	var cases []APITest
	cases = append(cases, APITest{BibleId: `ATIWBT`, Expected: 9747})
	APITestUtility(scriptTextScript, cases, t)
}

func TestScriptTextScript(t *testing.T) {
	var bibleId = `ATIWBT`
	ctx := context.Background()
	var req = strings.Replace(scriptTextScript, `{bibleId}`, bibleId, 2)
	var control = controller.NewController(ctx, []byte(req))
	filename, status := control.Process()
	if status != nil {
		t.Fatal(status)
	}
	fmt.Println("Filename:", filename)
	conn := db.NewDBAdapter(context.TODO(), filename)
	count, status := conn.CountScriptRows()
	if status != nil {
		t.Fatal(status)
	}
	var expected = 9747
	if count != expected {
		t.Error(`Expected `, expected, `records, got`, count)
	}
	identTest(`ScriptTextScript_`+bibleId, t, request.TextScript, ``,
		`ATIWBTN2ST`, ``, ``, `ati`)
}
