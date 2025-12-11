package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/controller"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
)

const uSXTextEditScript = `is_new: yes
dataset_name: 01c_usx_text_edit_{bibleId}
bible_id: {bibleId}
username: GaryNTest
testament:
  nt: yes
output:
  sqlite: yes
text_data:
  bible_brain:
    text_usx_edit: yes
detail:
  words: yes
`

func TestUSXTextEditDirect(t *testing.T) {
	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 8213})
	testName := strings.Replace(uSXTextEditScript, "{bibleId}", "ENGWEB", -1)
	DirectSqlTest(testName, tests, t)
}

// The tests below require json output

func TestUSXTextEditScriptAPI(t *testing.T) {
	var cases []APITest
	//cases = append(cases, APITest{BibleId: `ENGWEB`, Expected: 9588})
	//cases = append(cases, APITest{BibleId: `ATIWBT`, Expected: 8254})
	cases = append(cases, APITest{BibleId: `ABIWBT`, Expected: 8256})
	APITestUtility(uSXTextEditScript, cases, t)
}

func TestUSXTextEditScriptCLI(t *testing.T) {
	start := time.Now()
	var bibleId = `ENGWEB`
	var req = strings.Replace(uSXTextEditScript, `{bibleId}`, bibleId, 2)
	stdout, stderr := CLIExec(req, t)
	fmt.Println(`STDOUT:`, stdout)
	fmt.Println(`STDERR:`, stderr)
	fmt.Println("Duration:", time.Since(start))
	filename := ExtractFilename(req)
	numLines := NumJSONFileLines(filename, t)
	expected := 9588
	if numLines != expected {
		t.Error(`Expected `, expected, `records, got`, numLines)
	}
	identTest(`USX_Text_Edit_Script_`+bibleId, t, request.TextUSXEdit, ``,
		`ENGWEBN_ET-usx`, ``, ``, `eng`)
}

func TestUSXTextEditScript(t *testing.T) {
	var bibleId = `ATIWBT`
	ctx := context.Background()
	var req = strings.Replace(uSXTextEditScript, `{bibleId}`, bibleId, 2)
	var control = controller.NewController(ctx, []byte(req))
	filename, status := control.Process()
	if status != nil {
		t.Error(status)
	}
	fmt.Println(filename)
	numLines := NumJSONFileLines(filename, t)
	count := 8254
	if numLines != count {
		t.Error(`Expected `, count, `records, got`, numLines)
	}
	identTest(`USX_Text_Edit_Script_`+bibleId, t, request.TextUSXEdit, ``,
		`ATIWBTN_ET-usx`, ``, ``, `ati`)
}
