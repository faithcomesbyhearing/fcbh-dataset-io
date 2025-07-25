package tests

import (
	"testing"
)

const mmsAdapterASR = `is_new: no
dataset_name: 17a_mms_adapter
username: GaryNTest
language_iso: cul
notify_ok: [gary@shortsands.com]
notify_err: [gary@shortsands.com]
testament:
  nt_books: [PHM]
database:
  file: /Users/gary/FCBH2024/GaryNTest/17a_mms_adapter.db
audio_data:
  file: /Users/gary/FCBH2024/GaryNTest/17a_mms_adapter/*.mp3
speech_to_text:
  adapter_asr: y
compare:
  html_report: yes
  compare_settings: 
    lower_case: y
    remove_prompt_chars: y
    remove_punctuation: y
    double_quotes: 
      remove: y
    apostrophe: 
      remove: y
    hyphen:
      remove: y
    diacritical_marks:
      normalize_nfkd: y
`

func TestMMSAdapterASR(t *testing.T) {
	var yaml = mmsAdapterASR
	DirectSqlTest(yaml, []SqliteTest{}, t)
}
