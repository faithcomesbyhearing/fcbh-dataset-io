package tests

import (
	"strings"
	"testing"
)

const tSAeneasTest = `is_new: yes
dataset_name: 13b_ts_aeneas
bible_id: {bibleId}
username: GaryNTest
output:
  sqlite: yes
text_data:
  bible_brain:
    text_plain_edit: yes
audio_data:
  bible_brain:
    mp3_64: yes
timestamps:
  aeneas: yes
testament:
  nt_books: [1JN]
`

func TestTSAeneasDirect(t *testing.T) {
	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 110})
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts WHERE script_begin_ts != 0.0", 105})
	testName := strings.Replace(tSAeneasTest, "{bibleId}", "ENGWEB", -1)
	DirectSqlTest(testName, tests, t)
}

func TestPlainTextAeneasTimestampsScriptAPI(t *testing.T) {
	var tests []APITest
	tests = append(tests, APITest{BibleId: `ENGWEB`, Expected: 111, Diff: 0})
	tests = append(tests, APITest{BibleId: `ATIWBT`, Expected: 111, Diff: 0})
	APITestUtility(tSAeneasTest, tests, t)
}
