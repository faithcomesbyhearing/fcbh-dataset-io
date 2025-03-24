package tests

import (
	"strings"
	"testing"
)

const fastText = `is_new: yes
dataset_name: 12b_fasttext
bible_id: {bibleId}
username: GaryNTest
output:
  sqlite: yes
text_data:
  bible_brain:
    text_plain_edit: yes
text_encoding: 
  fast_text: yes
testament:
  nt: yes
detail:
  words: yes
`

func TestFasttextDirect(t *testing.T) {
	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 8218})
	tests = append(tests, SqliteTest{"SELECT count(*) FROM words WHERE ttype = 'W'", 175764})
	tests = append(tests, SqliteTest{"SELECT count(*) FROM words WHERE word_enc != ''", 175764})
	testName := strings.Replace(fastText, "{bibleId}", "ENGWEB", -1)
	DirectSqlTest(testName, tests, t)
}
