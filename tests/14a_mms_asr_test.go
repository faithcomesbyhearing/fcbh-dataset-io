package tests

import (
	"testing"
)

const mMSASRTest = `is_new: yes
dataset_name: 14a_mms_asr
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
  mms_align:
  bible_brain: yes
testament:
  nt_books: [PHM]
speech_to_text:
  mms_asr: yes
`

func TestMMSASRDirect(t *testing.T) {
	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 26})
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts WHERE script_begin_ts != 0.0", 25})
	DirectSqlTest(mMSASRTest, tests, t)
}
