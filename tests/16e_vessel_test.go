package tests

import (
	"testing"
)

const vesselTest = `is_new: yes
dataset_name: 16e_vessel_test
bible_id: spatit # incorrect
username: GaryNTest
testament:
  nt: yes
text_data:
  file: /Users/gary/FCBH2024/GaryNTest/16e_vessel_test.xlsx
audio_data:
  file: /Users/gary/FCBH2024/GaryNTest/16e_vessel_test/*_VOX.wav
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
      normalize_nfd: y
`

func TestVessel(t *testing.T) {
	var yaml = vesselTest
	DirectSqlTest(yaml, []SqliteTest{}, t)
}
