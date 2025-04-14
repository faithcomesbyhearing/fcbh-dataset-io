package tests

import (
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"testing"
)

const mMSASRCompare = `is_new: yes
dataset_name: 15a_mms_asr
bible_id: ENGWEB
username: GaryNTest
text_data:
  bible_brain:
    text_usx_edit: yes
audio_data:
  bible_brain:
    mp3_64: yes
timestamps:
  bible_brain: yes
testament:
  nt_books: ['3JN']
speech_to_text:
  mms_asr: yes
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
      normalize_nfd: y
`

func TestMMSASRCompare(t *testing.T) {
	var tests []CtlTest
	tests = append(tests, CtlTest{BibleId: "ENGWEB", Expected: 3, TextNtId: "ENGWEBN_ET-usx",
		TextType: request.TextUSXEdit, AudioNTId: "ENGWEBN2DA", Language: "eng"})
	//tests = append(tests, CtlTest{BibleId: "APFCMU", Expected: 16, TextNtId: "APFCMUN_ET-usx",
	//	AudioNTId: `APFCMUN1DA`, TextType: request.TextUSXEdit, Language: "apf"})
	DirectTestUtility(mMSASRCompare, tests, t)
}
