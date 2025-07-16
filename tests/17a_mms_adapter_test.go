package tests

import (
	"testing"
)

const mmsAdapter = `is_new: no
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
training:
  mms_adapter:
    batch_mb: 4
    num_epochs: 1
    learning_rate: 3e-5
    warmup_pct: 3.0
    grad_norm_max: 1.0
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

func TestMMSAdapter(t *testing.T) {
	var yaml = mmsAdapter
	DirectSqlTest(yaml, []SqliteTest{}, t)
}
