package decode_yaml

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	var d = NewRequestDecoder(context.Background())
	content, err := os.ReadFile(`request_test.yaml`)
	if err != nil {
		panic(err)
	}
	var req, _ = d.Decode(content)
	req.IsNew = true
	//req.BibleId = `EBGESV`
	req.LanguageISO = `eng`
	req.AltLanguage = `en`
	req.NotifyOk = []string{"abc"}
	req.NotifyErr = []string{"def"}
	req.AudioData.File = `file:///where`
	req.AudioData.BibleBrain.MP3_64 = false
	req.AudioData.BibleBrain.OPUS = false
	req.AudioData.POST = `` //`filename`
	req.TextData.NoText = false
	req.TextData.BibleBrain.TextPlain = true
	req.Timestamps.MMSFAVerse = true
	//req.Timestamps.MMSFAWord = true
	//req.SpeechToText.Whisper.Model.Medium = true
	req.SpeechToText.MMS = true
	req.AudioEncoding.MFCC = true
	req.AudioEncoding.NoEncoding = false
	req.Compare.CompareSettings.Apostrophe.Normalize = true
	req.Compare.CompareSettings.Apostrophe.Remove = false
	d.Validate(&req)
	d.Prereq(&req)
	d.Depend(req)
	if len(d.errors) > 0 {
		t.Fatal(strings.Join(d.errors, "\n"))
	}
}

// I should have a test with multiple error
// I shoud have a test with one selected, not error
// I should have a test with none selected
