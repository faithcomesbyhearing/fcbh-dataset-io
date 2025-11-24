package uroman

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
)

func TestURoman(t *testing.T) {
	ctx := context.Background()
	var input = []string{"Игорь Стравинский",
		"Игорь",
		"Ντέιβις Καπ",
		"\u0041",
		"123 \u09E6\u09EF \u0966-\u096F \u0660-\u0669", // numbers changed to ascii
		"comma: \u3001 period: \u3002 corners: \u300C\u300D reference: \u203B middle dot: \u30FB",
		"closing double quote: \u030B inverted caret: \u030C upper right: \u031A", // diacriticals are ignored
		"NFC: \u00FC NFD: u\u0308", // NFC processed, NFD diacritical ignored
		"क\u0947", "क", "\u0947",
	}
	var scripts []db.Script
	for i, in := range input {
		var rec db.Script
		rec.ScriptId = i + 1
		rec.BookId = "RUS"
		rec.ChapterNum = 1
		rec.VerseStr = strconv.Itoa(rec.ScriptId)
		rec.ScriptTexts = append(rec.ScriptTexts, in)
		scripts = append(scripts, rec)
	}
	conn := db.NewDBAdapter(ctx, ":memory:")
	status := conn.InsertScripts(scripts)
	if status != nil {
		t.Fatal(status)
	}
	rows, status := conn.SelectScripts()
	for _, r := range rows {
		fmt.Println(r.ScriptId, r.ScriptText, r.URoman)
	}
	status = EnsureUroman(conn, "rus")
	if status != nil {
		t.Fatal(status)
	}
	var hasUroman bool
	hasUroman, status = CheckUroman(conn)
	if status != nil {
		t.Fatal(status)
	}
	if !hasUroman {
		t.Error("Uroman has not updated the database")
	}
	results, status := conn.SelectScripts()
	for _, line := range results {
		fmt.Println(line.URoman)
	}
	if len(results) != len(scripts) {
		t.Error("Result rows should equal script rows", len(results), len(scripts))
	}
}
