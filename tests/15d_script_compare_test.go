package tests

import (
	"testing"
)

const scriptCompare = `is_new: no
dataset_name: 15d_script_compare_audio
bible_id: IKHMLT
username: GaryNTest
email: gary@shortsands.com
output:
  sqlite: yes
testament:
  nt: yes
audio_proof:
  html_report: no # must eliminate directory input files passed into module to do this
  base_dataset: 15d_script_compare
compare:
  html_report: yes
  base_dataset: 15d_script_compare
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

func TestScriptCompare(t *testing.T) {
	DirectSqlTest(scriptCompare, []SqliteTest{}, t)
}
