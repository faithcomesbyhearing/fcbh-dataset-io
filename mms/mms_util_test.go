package mms

import (
	"context"
	"testing"
)

func TestMMSUtil_checkLanguagea(t *testing.T) {
	checkLanguageTest("fra", "en", "en", t)
	checkLanguageTest("fra", "", "fra", t)
	checkLanguageTest("npi", "", "npi", t)
	checkLanguageTest("abc", "", "pam", t)
}

func checkLanguageTest(lang string, sttLang string, expectLang string, t *testing.T) {
	ctx := context.Background()
	resultLang, status := CheckLanguage(ctx, lang, sttLang, "mms_asr")
	if status != nil {
		t.Error(status)
	}
	if resultLang != expectLang {
		t.Error("actual:", resultLang, "expected:", expectLang)
	}
}
