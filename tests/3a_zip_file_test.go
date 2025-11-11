package tests

import (
	"testing"
)

const zipFile = `is_new: yes
dataset_name: 3a_zip_file
bible_id: ENGWEB
username: GaryNTest
output:
  sqlite: yes
audio_data:
  file: /Users/gary/FCBH2024/download/ENGWEBN2DA-mp3-64.zip
text_data:
  file: /Users/gary/FCBH2024/download/ENGWEBN_ET-usx.zip
testament:
  nt_books: [MRK]
`

func TestZipFileDirect(t *testing.T) {
	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 694})
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts WHERE script_begin_ts != 0.0", 0})
	DirectSqlTest(zipFile, tests, t)
}
