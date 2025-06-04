package decode_yaml

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestRequestYamlFile(t *testing.T) {
	var d = NewRequestDecoder(context.Background())
	content, err := os.ReadFile(`request_test.yaml`)
	if err != nil {
		panic(err)
	}
	req, status := d.Decode(content)
	if status != nil {
		t.Fatal(status)
	}
	yaml, status := d.Encode(req)
	if status != nil {
		t.Fatal(status)
	}
	fmt.Println(yaml)
}

func TestParser(t *testing.T) {
	var test1 = `is_new: Yes
dataset_name: Test1  # should be a unique name
username: GaryNGriswold
#bible_id: ENGWEB
language_iso: eng
alt_language: en
notify_ok: [emille@fcbhmail.org]
notify_err: [jbarndt@fcbhmail.org]
output:
  directory: /Users/gary
  csv: Yes
  json: Yes
  sqlite: Yes
testament:
  ot: 
  ot_books: [GEN,EXO,LEV,NUM,DEU]
  nt: Yes
  nt_books: []
audio_data:
  bible_brain:
    mp3_64: Yes
    mp3_16: Y
    opus: y
  file: file:///Users/who/where.mp3
  aws_s3: http://west1/path
  post: Yes
  no_audio: Yes
text_data: 
  bible_brain: 
    text_usx_edit: Yes
    text_plain_edit: Yes
    text_plain: Yes
  file: /Users/who/where.mp3
  aws_s3: http://west1/path
  post: Yes
  no_text: Yes
timestamps:
  bible_brain: yes
  aeneas: yes
  ts_bucket: yes
  mms_fa_verse: yes
  mms_align: yes
  no_timestamps: yes
speech_to_text:
  mms_asr: Yes
  whisper:
    model: 
      large: Yes
      medium: Yes
      small: Yes
      base: Yes
      tiny: Yes
  no_speech_to_text: Yes
detail:
  lines: yes
  words: yes
audio_encoding:
  mfcc: yes
  no_encoding: yes
text_encoding:
  fast_text: yes
  no_encoding: yes
audio_proof:
  html_report: yes
  base_dataset: UseProject1
compare:
  html_report: yes
  base_dataset: UseProject1
  compare_settings:
    lower_case: yes
    remove_prompt_chars: yes
    remove_punctuation: yes
    double_quotes: 
      remove: yes
      normalize: yes
    apostrophe: 
      remove: yes
      normalize: yes
    hyphen: 
      remove: yes
      normalize: yes
    diacritical_marks: 
      remove: yes
      normalize_nfc: yes
      normalize_nfd: yes
      normalize_nfkc: yes
      normalize_nfkd: yes`
	var r = NewRequestDecoder(context.Background())
	req, status := r.Decode([]byte(test1))
	if status != nil {
		t.Fatal(status)
	}
	if !req.TextEncoding.FastText {
		t.Error("FastText should be true")
	}
	_, _ = r.Encode(req)
	var boolTests = []bool{
		req.Output.CSV,
		req.Output.JSON,
		req.Output.Sqlite,
		//req.Testament.OT,
		req.Testament.NT,
		req.AudioData.BibleBrain.MP3_64,
		req.AudioData.BibleBrain.MP3_16,
		req.AudioData.BibleBrain.OPUS,
		req.AudioData.NoAudio,
		req.TextData.BibleBrain.TextUSXEdit,
		req.TextData.BibleBrain.TextPlainEdit,
		req.TextData.BibleBrain.TextPlain,
		req.TextData.NoText,
		req.Timestamps.BibleBrain,
		req.Timestamps.Aeneas,
		req.Timestamps.TSBucket,
		req.Timestamps.MMSFAVerse,
		req.Timestamps.MMSAlign,
		req.Timestamps.NoTimestamps,
		req.SpeechToText.MMS,
		req.SpeechToText.Whisper.Model.Large,
		req.SpeechToText.Whisper.Model.Medium,
		req.SpeechToText.Whisper.Model.Small,
		req.SpeechToText.Whisper.Model.Base,
		req.SpeechToText.Whisper.Model.Tiny,
		req.SpeechToText.NoSpeechToText,
		req.Detail.Lines,
		req.Detail.Words,
		req.AudioEncoding.MFCC,
		req.AudioEncoding.NoEncoding,
		req.TextEncoding.FastText,
		req.TextEncoding.NoEncoding,
		req.AudioProof.HTMLReport,
		req.Compare.HTMLReport,
		req.Compare.CompareSettings.LowerCase,
		req.Compare.CompareSettings.RemovePromptChars,
		req.Compare.CompareSettings.RemovePunctuation,
		req.Compare.CompareSettings.DoubleQuotes.Remove,
		req.Compare.CompareSettings.DoubleQuotes.Normalize,
		req.Compare.CompareSettings.Apostrophe.Remove,
		req.Compare.CompareSettings.Apostrophe.Normalize,
		req.Compare.CompareSettings.Hyphen.Remove,
		req.Compare.CompareSettings.Hyphen.Normalize,
		req.Compare.CompareSettings.DiacriticalMarks.Remove,
		req.Compare.CompareSettings.DiacriticalMarks.NormalizeNFC,
		req.Compare.CompareSettings.DiacriticalMarks.NormalizeNFD,
		req.Compare.CompareSettings.DiacriticalMarks.NormalizeNFKC,
		req.Compare.CompareSettings.DiacriticalMarks.NormalizeNFKD}
	for i, item := range boolTests {
		if !item {
			t.Error(`The`, i, `th item should be true, but is not`)
		}
	}
	var strs = []string{req.DatasetName,
		req.BibleId,
		req.Username,
		req.Output.Directory,
		req.AudioData.File,
		req.AudioData.AWSS3,
		req.TextData.File,
		req.TextData.AWSS3,
		req.Compare.BaseDataset}
	for i, item := range strs {
		if len(item) == 0 {
			t.Error(`The`, i, `th item should have a value, but is empty`)
		}
	}
	if len(req.Testament.OTBooks) != 5 {
		t.Error(`OTBooks should have a length of 5, not`, len(req.Testament.OTBooks))
	}
	if len(req.Testament.NTBooks) != 0 {
		t.Error(`NTBooks should have a length of 0, not`, len(req.Testament.NTBooks))
	}
	if len(req.NotifyOk) != 1 {
		t.Error(`NotifyOk should have a length of 1, not`, len(req.NotifyOk))
	}
	if len(req.NotifyErr) != 1 {
		t.Error(`NotifyErr should have a length of 1, not`, len(req.NotifyErr))
	}
}
