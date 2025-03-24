package tests

import (
	"testing"
)

const proofCompare = `is_new: no
dataset_name: 15c_proof_compare_audio
bible_id: ENGWEB
username: GaryNTest
notify_ok: [gary@shortsands.com]
notify_err: [gary@shortsands.com]
testament:
  nt: yes
audio_data:
  bible_brain:
    mp3_64: yes
audio_proof:
  html_report: yes
  base_dataset: 15c_proof_compare
compare:
  html_report: yes
  base_dataset: 15c_proof_compare
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

// TestProofAndCompare is not compatible with the debugger, but puts results in bucket
func TestProofAndCompare(t *testing.T) {
	CLIExec(proofCompare, t)
}

// TestProofAndCompareWithDebugger does not put output in the bucket
func TestProofAndCompareWithDebugger(t *testing.T) {
	DirectSqlTest(proofCompare, []SqliteTest{}, t)
}
