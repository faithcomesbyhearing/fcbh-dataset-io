package tests

import "testing"

const runAnything = `is_new: yes
dataset_name: ART
language_iso: spa
username: GaryNTest
notify_ok: [gary@shortsands.com]
notify_err: [gary@shortsands.com]
text_data:
  aws_s3: s3://dataset-vessel/vessel/ART_12231842/ART Text/XLSX/Arti Test_ART_LineBased.xlsx
audio_data:
  aws_s3: s3://dataset-vessel/vessel/ART_12231842/ART Line VOX/*.wav
timestamps:
  mms_align: y
training:
  redo_training: yes
  mms_adapter:
    batch_mb: 4
    num_epochs: 16
    learning_rate: 1e-3
    warmup_pct: 12.0
    grad_norm_max: 0.4
speech_to_text:
  adapter_asr: yes
compare:
  html_report: yes
  gordon_filter: 0
  compare_settings:
    lower_case: yes
    remove_prompt_chars: yes
    remove_punctuation: yes
    double_quotes:
      remove: yes
    apostrophe:
      remove: yes
    hyphen:
      remove: yes
    diacritical_marks:
      normalize_nfc: yes
`

func TestRunAnything(t *testing.T) {
	var yaml = runAnything
	DirectSqlTest(yaml, []SqliteTest{}, t)
}
