package tests

import (
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"strings"
	"testing"
)

const plainTextScript = `is_new: yes
dataset_name: 01a_plain_text_{bibleId}
bible_id: {bibleId}
username: GaryNTest
output:
  sqlite: yes
text_data:
  bible_brain:
    text_plain: yes
detail:
  words: yes
`

func TestPlainTextDirect(t *testing.T) {
	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 7958})
	testName := strings.Replace(plainTextScript, "{bibleId}", "ENGWEB", -1)
	DirectSqlTest(testName, tests, t)
}

// The tests below need csv output to work

func TestPlainTextScriptAPI(t *testing.T) {
	var cases []APITest
	cases = append(cases, APITest{BibleId: `ENGWEB`, Expected: 7959, Diff: 0})
	APITestUtility(plainTextScript, cases, t)
}

func TestPlainTextScriptCLI(t *testing.T) {
	var bibleId = `ENGWEB`
	var req = strings.Replace(plainTextScript, `{bibleId}`, bibleId, -1)
	stdout, stderr := CLIExec(req, t)
	fmt.Println(`STDOUT:`, stdout)
	fmt.Println(`STDERR:`, stderr)
	filename := ExtractFilename(req)
	numLines := NumCVSFileLines(filename, t)
	count := 7959
	if numLines != count {
		t.Error(`Expected `, count, `records, got`, numLines)
	}
	identTest(`PlainTextScript_`+bibleId, t, request.TextPlain, ``,
		`ENGWEBN_ET`, ``, ``, `eng`)
}
