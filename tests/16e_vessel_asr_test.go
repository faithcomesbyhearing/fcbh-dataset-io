package tests

import (
	"testing"

	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

const vesselASRTest = `is_new: yes
dataset_name: 16e_vessel_test
username: GaryNTest
language_iso: eng
notify_ok: [gary@shortsands.com]
notify_err: [gary@shortsands.com]
testament:
  nt: yes
  ot: yes
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

func TestVesselASR(t *testing.T) {
	log.SetOutput("stderr")
	var yaml = vesselASRTest
	DirectSqlTest(yaml, []SqliteTest{}, t)
}
