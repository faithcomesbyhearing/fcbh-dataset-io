package tests

import (
	"testing"
)

const mmsAlignTest = `is_new: yes
dataset_name: 13e_mms_align
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
  mms_align: yes
testament:
  nt_books: [PHM]
`

// These counts are 2 words short  and 11 letters short, because I removed
// the chapter heading from the first chapter of each book.
// When I restore that the counts will need to be increased.

func TestMMSAlignDirect(t *testing.T) {
	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 26})
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts WHERE script_begin_ts != 0.0", 25})
	tests = append(tests, SqliteTest{"SELECT count(*) FROM words WHERE ttype='W'", 446})
	tests = append(tests, SqliteTest{"SELECT count(*) FROM words WHERE word_begin_ts != 0.0", 446})
	tests = append(tests, SqliteTest{"SELECT count(*) FROM chars", 1785})
	tests = append(tests, SqliteTest{"SELECT count(distinct(word_id)) FROM chars", 446})
	DirectSqlTest(mmsAlignTest, tests, t)
}
