package tests

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/generic"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const usxVsPlain = `is_new: yes
dataset_name: 3b_usx_vs_plain_{BIBLE_ID}
bible_id: {BIBLE_ID}
username: GaryNTest
testament:
  nt: yes
output:
  sqlite: yes
text_data:
  bible_brain:
    {TEXT_TYPE}: yes
detail:
  words: yes
`

func TestTextReadDirect(t *testing.T) {
	var tests1 []SqliteTest
	tests1 = append(tests1, SqliteTest{"SELECT count(*) FROM scripts", 8213})
	tests1 = append(tests1, SqliteTest{"SELECT count(*) FROM words where ttype='W'", 175764})
	test1 := strings.ReplaceAll(usxVsPlain, "{BIBLE_ID}", "ENGWEB")
	test1USX := strings.ReplaceAll(test1, "{TEXT_TYPE}", "text_usx_edit")
	database1 := DirectSqlTest(test1USX, tests1, t)
	var tests2 []SqliteTest
	tests2 = append(tests2, SqliteTest{"SELECT count(*) FROM scripts", 8218})
	tests2 = append(tests2, SqliteTest{"SELECT count(*) FROM words where ttype='W'", 175764})
	text2TXT := strings.ReplaceAll(test1, "{TEXT_TYPE}", "text_plain_edit")
	database2 := DirectSqlTest(text2TXT, tests2, t)
	diffCount := DifferenceTest(database1, database2)
	if diffCount != 0 {
		t.Error("DiffCount is expected to be 0, but was", diffCount)
	}
}

func DifferenceTest(database1 string, database2 string) int {
	ctx := context.Background()
	diffMatch := diffmatchpatch.New()
	conn1 := db.NewDBAdapter(ctx, "./"+database1)
	records1, status := conn1.SelectScripts()
	if status != nil {
		panic(status)
	}
	var usxMap = make(map[string]string)
	for _, rec := range records1 {
		var lf generic.VerseRef
		lf.BookId = rec.BookId
		lf.ChapterNum = rec.ChapterNum
		lf.VerseStr = rec.VerseStr
		usxMap[lf.UniqueKey()] = rec.ScriptText
	}
	conn1.Close()
	conn2 := db.NewDBAdapter(ctx, "./"+database2)
	plainRec2, _ := conn2.SelectScripts()
	var diffCount = 0
	for _, rec := range plainRec2 {
		var lf generic.VerseRef
		lf.BookId = rec.BookId
		lf.ChapterNum = rec.ChapterNum
		lf.VerseStr = rec.VerseStr
		lineRef := lf.UniqueKey()
		usxTxt, ok := usxMap[lineRef]
		if !ok {
			usxTxt = ""
		}
		usxTxt = strings.TrimSpace(usxTxt)
		plainTxt := rec.ScriptText
		re := regexp.MustCompile(`\n\s+`)
		plainTxt = re.ReplaceAllString(plainTxt, " ")
		usxTxt = strings.ReplaceAll(usxTxt, "  ", " ")
		plainTxt = strings.ReplaceAll(plainTxt, "   ", " ")
		plainTxt = strings.ReplaceAll(plainTxt, "  ", " ")
		diffs := diffMatch.DiffMain(usxTxt, plainTxt, false)
		diffs = diffMatch.DiffCleanupMerge(diffs)
		if len(diffs) > 1 {
			diffs = removeWhitespaceErrors(diffs)
			diffs = diffMatch.DiffCleanupMerge(diffs)
		}
		if len(diffs) > 1 || len(diffs) > 0 && diffs[0].Type != diffmatchpatch.DiffEqual {
			diffCount++
			fmt.Println(lineRef, "usxTxt", usxTxt)
			fmt.Println(lineRef, "plnTxt", plainTxt)
			fmt.Println(lineRef, diffMatch.DiffPrettyText(diffs))
			fmt.Println(lineRef, diffs)
			fmt.Println()
		}
	}
	fmt.Println("DiffCount", diffCount)
	return diffCount
}

func removeWhitespaceErrors(diffs []diffmatchpatch.Diff) []diffmatchpatch.Diff {
	var cleanDiffs []diffmatchpatch.Diff
	unEqualCount := 0
	for _, d := range diffs {
		if d.Type == diffmatchpatch.DiffEqual {
			cleanDiffs = append(cleanDiffs, d)
		} else if len(strings.TrimSpace(d.Text)) > 0 {
			cleanDiffs = append(cleanDiffs, d)
			unEqualCount++
		}
	}
	return cleanDiffs
}
