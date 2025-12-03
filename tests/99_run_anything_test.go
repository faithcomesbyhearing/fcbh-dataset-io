package tests

import "testing"

const runAnything = `is_new: yes 
dataset_name: N2MZJSIM
language_iso: mzj
username: GaryNTest 
notify_ok: [gary@shortsands.com]
notify_err: [gary@shortsands.com]
testament:
  nt_books: [3JN]
text_data:
  aws_s3: s3://pretest-audio/N2MZJSIM Manya (MZJ)/N2MZJSIM USX/*.usx
audio_data:
  aws_s3: s3://pretest-audio/N2MZJSIM Manya (MZJ)/N2MZJSIM Chapter VOX/*.mp3
speech_to_text:
  mms_asr_align: yes
compare:
  html_report: yes
  gordon_filter: 4
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
output:
  csv: y
`

func TestRunAnything(t *testing.T) {
	var yaml = runAnything
	DirectSqlTest(yaml, []SqliteTest{}, t)
}
