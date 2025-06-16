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
#testament:
#  nt_books: [PHM]
database:
  file: /Users/gary/FCBH2024/GaryNTest/17a_mms_adapter.db
audio_data:
  file: /Users/gary/FCBH2024/GaryNTest/17a_mms_adapter/*.mp3
training:
  mms_adapter:
    batch_size: 1
    num_epochs: 1
speech_to_text:
  mms_asr: y
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
      normalize_nfkd: y
`

func TestMMSAdapter(t *testing.T) {
	var yaml = mmsAdapter
	DirectSqlTest(yaml, []SqliteTest{}, t)
}

const mmsAdapterProd = `is_new: no
dataset_name: 17a_mms_adapter
username: GaryNTest
language_iso: cul
notify_ok: [gary@shortsands.com]
notify_err: [gary@shortsands.com]
testament:
  nt: y
database:
  file: /home/ec2-user/data/GaryNTest/N2CUL_MNT_input.db
audio_data:
  aws_s3: s3://pretest-audio/N2CULMNT Kulina (CUL)/CULMNTN2DA/*.mp3
training:
  mms_adapter:
    batch_size: 1
    num_epochs: 1
speech_to_text:
  mms_asr: y
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
      normalize_nfkd: y
`
