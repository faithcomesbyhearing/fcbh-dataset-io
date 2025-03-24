package tests

import (
	"testing"
)

const fAVerseTest = `is_new: yes
dataset_name: 13d_fa_verse
bible_id: ENGWEB
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
  mms_fa_verse: yes
testament:
  nt_books: [PHM]
`

func TestFAVerseDirect(t *testing.T) {
	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 26})
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts WHERE script_begin_ts != 0.0", 25})
	DirectSqlTest(fAVerseTest, tests, t)
}
