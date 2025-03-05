package tests

import (
	"testing"
)

const scriptCompare = `is_new: yes
dataset_name: N2IKHMLT
bible_id: IKHMLT
username: GaryNTest
email: gary@shortsands.com
output:
  sqlite: yes
testament:
  nt: yes 
text_data:
  file: /Users/gary/FCBH2024/download/IKHMLT/Vessel Text_Aokho_N2IKHMLT.xlsx
audio_data:
  aws_s3: s3://pretest-audio/N2IKHMLT Aokho (IKH)/N2IKHMLT Chapter VOX/*.mp3
timestamps:
  mms_align: no
speech_to_text:
  mms_asr: no
audio_proof:
  html_report: no
compare:
  html_report: no
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
