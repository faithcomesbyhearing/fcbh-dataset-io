package diff

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"os"
	"path/filepath"
	"testing"
)

type compareTest struct {
	baseDB  string
	project string
	expect  int
}

func TestCompare(t *testing.T) {
	var tests []compareTest
	//tests = append(tests, testCase{baseDB: "ATIWBT_SCRIPT.db", project: "ATIWBT_USXEDIT", expect: 2})
	//tests = append(tests, testCase{baseDB: `ENGWEB_WHISPER.db`, project: `ENGWEB_WHISPER_STT`, expect: 90})
	//tests = append(tests, testCase{baseDB: `APFCMU_WHISPER.db`, project: `APFCMU_WHISPER_STT`, expect: 91})
	//tests = append(tests, testCase{baseDB: `DYIIBS_WHISPER.db`, project: `DYIIBS_WHISPER_STT`, expect: 91})
	tests = append(tests, compareTest{baseDB: "N2ENGWEB", project: "N2ENGWEB_audio", expect: 2479})

	for _, tst := range tests {
		records, _, status := runCompareTest(tst)
		fmt.Println(status, len(records))
		if len(records) != tst.expect {
			t.Error(`Expected count is`, tst.expect, ` actual was`, len(records))
		}
		//if len(fileMap) != tst.expect {}
	}
}

func runCompareTest(tst compareTest) ([]Pair, string, *log.Status) {
	var records []Pair
	var fileMap string
	ctx := context.Background()
	user := ``
	var testament = request.Testament{NT: true}
	var cfg request.CompareSettings
	cfg.LowerCase = true
	cfg.RemovePromptChars = true
	cfg.RemovePunctuation = true
	cfg.DoubleQuotes.Remove = true
	cfg.Apostrophe.Remove = true
	cfg.Hyphen.Remove = true
	cfg.DiacriticalMarks.NormalizeNFD = true
	_ = os.Setenv("FCBH_DATASET_DB", filepath.Join(os.Getenv("GOPROJ"), "match"))
	conn, status := db.NewerDBAdapter(ctx, false, user, tst.baseDB)
	if status != nil {
		return records, fileMap, status
	}
	compare := NewCompare(ctx, user, tst.project, conn, "eng", testament, cfg)
	return compare.Process()
}
