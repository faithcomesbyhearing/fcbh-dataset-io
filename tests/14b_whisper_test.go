package tests

import (
	"testing"
)

const whisperTest = `is_new: yes
dataset_name: 14b_whisper
bible_id: ENGWEB
username: GaryNTest
output:
  sqlite: yes
testament: 
  nt_books: [PHM]
text_data:
  bible_brain:
    text_usx_edit: yes
audio_data:
  bible_brain:
    mp3_64: yes
timestamps:
  bible_brain: yes
speech_to_text:
  whisper:
    model:
      tiny: yes
`

func TestWhisperDirect(t *testing.T) {
	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 26})
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts WHERE script_begin_ts != 0.0", 25})
	DirectSqlTest(whisperTest, tests, t)
}
