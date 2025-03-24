package tests

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/controller"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"strings"
	"testing"
	"time"
)

const plainTextEditScript = `is_new: yes
dataset_name: 01b_plain_text_edit_{bibleId}
bible_id: {bibleId}
username: GaryNTest
output:
  sqlite: yes
text_data:
  bible_brain:
    text_plain_edit: yes
detail:
  words: yes
`

func TestPlainTextEditDirect(t *testing.T) {
	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 8218})
	testName := strings.Replace(plainTextEditScript, "{bibleId}", "ENGWEB", -1)
	DirectSqlTest(testName, tests, t)
}

// The tests below require json output

func TestPlainTextEditScriptAPI(t *testing.T) {
	var cases []APITest
	cases = append(cases, APITest{BibleId: `ENGWEB`, Expected: 8218})
	cases = append(cases, APITest{BibleId: `ATIWBT`, Expected: 8216})
	APITestUtility(plainTextEditScript, cases, t)
}

func TestPlainTextEditScriptCLI(t *testing.T) {
	var start = time.Now()
	var bibleId = `ENGWEB`
	//var expected = 8218 // when detail = lines
	var expected = 175829 // when detail = words
	var req = strings.Replace(plainTextEditScript, `{bibleId}`, bibleId, 2)
	stdout, stderr := CLIExec(req, t)
	fmt.Println(`STDOUT:`, stdout)
	fmt.Println(`STDERR:`, stderr)
	fmt.Println("Duration:", time.Since(start))
	filename := ExtractFilename(req)
	numLines := NumJSONFileLines(filename, t)
	if numLines != expected {
		t.Error(`Expected `, expected, `records, got`, numLines)
	}
	identTest(`PlainTextEditScript_`+bibleId, t, request.TextPlainEdit, ``,
		`ENGWEBN_ET`, ``, ``, `eng`)
}

func TestPlainTextEditScript(t *testing.T) {
	type test struct {
		bibleId  string
		expected int
		textNtId string
		language string
	}
	var tests []test
	tests = append(tests, test{bibleId: "ENGWEB", expected: 8218, textNtId: "ENGWEBN_ET", language: "eng"})
	tests = append(tests, test{bibleId: "ATIWBT", expected: 8216, textNtId: "ATIWBTN_ET", language: "ati"})
	ctx := context.Background()
	for _, tst := range tests {
		var req = strings.Replace(plainTextEditScript, `{bibleId}`, tst.bibleId, 2)
		var control = controller.NewController(ctx, []byte(req))
		filename, status := control.Process()
		if status != nil {
			t.Error(status)
		}
		fmt.Println(filename)
		numLines := NumJSONFileLines(filename, t)
		if numLines != tst.expected {
			t.Error(`Expected `, tst.expected, `records, got`, numLines)
		}
		identTest(`PlainTextEditScript_`+tst.bibleId, t, request.TextPlainEdit, ``,
			tst.textNtId, ``, ``, tst.language)
	}
}
