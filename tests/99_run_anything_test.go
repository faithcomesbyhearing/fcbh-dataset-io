package tests

import "testing"

const runAnything = `is_new: yes
dataset_name: 10102401
language_iso: qqq
username: vessel
notify_ok: [gary@shortsands.com]
notify_err: [gary@shortsands.com]
testament:
  nt: yes
text_data:
  aws_s3: s3://dataset-vessel/vessel/10102401_06241918/10102401 Text/XLSX/Chapter Review_10102401_LineBased.xlsx
audio_data:
  aws_s3: s3://dataset-vessel/vessel/10102401_06241918/10102401 Line VOX/*.wav
speech_to_text:
  mms_asr: yes
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
      normalize_nfc: y
`

func TestRunAnything(t *testing.T) {
	var yaml = runAnything
	DirectSqlTest(yaml, []SqliteTest{}, t)
}
