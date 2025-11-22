package tests

import (
	"testing"

	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

// This is not a test that is expected to run to completion.
// It exists so that one can debug the initial parts of training
// Monitor the process on $FCBH_DATASET_DB/dataset.log
// This template says nt_books: [PHM], but I don't think the training module has the ability

const mmsAdapterASR = `is_new: no
dataset_name: 17a_mms_adapter
username: GaryNTest
language_iso: cul
notify_ok: [gary@shortsands.com]
notify_err: [gary@shortsands.com]
testament:
  nt_books: [PHM]
#database:
#  file: /Users/gary/FCBH2024/GaryNTest/17a_mms_adapter.db
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
	log.SetOutput("stderr")
	var yaml = mmsAdapterASR
	DirectSqlTest(yaml, []SqliteTest{}, t)
}
