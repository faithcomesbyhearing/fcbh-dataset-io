package tests

import "testing"

const runAnything = `is_new: no 
dataset_name: N2KEUWB4
bible_id: KEUWB4
username: GaryNTest 
notify_ok: [gary@shortsands.com]
notify_err: [gary@shortsands.com]
testament:
  nt_books: [LUK,ROM,1JN]
#text_data:
#  aws_s3: s3://pretest-audio/N2KEUWB4 Akebu (BOV)/N2KEUWBT Text/USX/*.usx
audio_data:
  aws_s3: s3://pretest-audio/N2KEUWB4 Akebu (BOV)/N2KEUWBT Chapter VOX/*.mp3
#timestamps:
#  mms_align: yes
training:
  redo_training: no
  wav2vec2_word:
    batch_mb: 4
    num_epochs: 1 
    learning_rate: 7.5e-5 # 5e-5 to 1e-4 suggested
    warmup_pct: 12.0 # 10-15
    grad_norm_max: 3.0 # 1-5
    min_audio_sec: 0.5
speech_to_text:
  wav2vec2_asr: yes
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
`

func TestRunAnything(t *testing.T) {
	var yaml = runAnything
	DirectSqlTest(yaml, []SqliteTest{}, t)
}
