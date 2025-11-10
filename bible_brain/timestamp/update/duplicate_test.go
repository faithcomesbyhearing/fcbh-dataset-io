package update

import (
	"os"
	"testing"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
)

func TestCompareDurationsAllMatch(t *testing.T) {
	source := map[string]map[int]float64{
		"GEN": {1: 120.0, 2: 98.5},
	}
	target := map[string]map[int]float64{
		"GEN": {1: 120.0, 2: 98.5},
	}

	result := compareDurations(source, target, 0)
	if len(result.Mismatches) != 0 {
		t.Fatalf("expected no mismatches, got %d", len(result.Mismatches))
	}
	if len(result.Chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(result.Chapters))
	}
}

func TestCompareDurationsMismatch(t *testing.T) {
	source := map[string]map[int]float64{
		"GEN": {1: 120.0},
	}
	target := map[string]map[int]float64{
		"GEN": {1: 110.0},
	}

	result := compareDurations(source, target, 0)
	if len(result.Mismatches) != 1 {
		t.Fatalf("expected 1 mismatch, got %d", len(result.Mismatches))
	}
}

func TestFilterTimestampData(t *testing.T) {
	data := map[string]map[int][]Timestamp{
		"GEN": {
			1: {{VerseStr: "1"}},
			2: {{VerseStr: "2"}},
		},
	}
	existing := []db.Script{{BookId: "GEN", ChapterNum: 1}, {BookId: "GEN", ChapterNum: 2}}
	allowed := []db.Script{{BookId: "GEN", ChapterNum: 2}}

	filtered, chapters := filterTimestampData(data, existing, allowed)
	if len(chapters) != 1 || chapters[0].ChapterNum != 2 {
		t.Fatalf("expected chapter 2, got %+v", chapters)
	}
	if len(filtered["GEN"]) != 1 {
		t.Fatalf("expected filtered map to only include one chapter")
	}
}

func TestDuplicationTolerance(t *testing.T) {
	os.Unsetenv("BB_DUPLICATION_TOLERANCE")
	if val := duplicationTolerance(); val != 0 {
		t.Fatalf("expected default 0 tolerance")
	}

	os.Setenv("BB_DUPLICATION_TOLERANCE", "0.75")
	if val := duplicationTolerance(); val != 0.75 {
		t.Fatalf("expected 0.75 tolerance, got %f", val)
	}
	os.Unsetenv("BB_DUPLICATION_TOLERANCE")
}

func TestInferSourceFileset(t *testing.T) {
	ident := db.Ident{AudioNTId: "ABCDEN1DA", AudioOTId: "ABCDOO1DA"}

	if src := inferSourceFileset(ident, "ABCDEFGN2DA"); src != "ABCDEN1DA" {
		t.Fatalf("expected NT source, got %s", src)
	}

	if src := inferSourceFileset(ident, "ABCDEFGO2DA"); src != "ABCDOO1DA" {
		t.Fatalf("expected OT source, got %s", src)
	}

	if src := inferSourceFileset(ident, "ABCDEFGN1DA"); src != "" {
		t.Fatalf("expected no source for N1 target")
	}
}

func TestResetTimestampIDs(t *testing.T) {
	data := map[string]map[int][]Timestamp{
		"GEN": {
			1: {
				{TimestampId: 12},
				{TimestampId: 34},
			},
		},
	}

	resetTimestampIDs(data)

	for _, chapterMap := range data {
		for _, list := range chapterMap {
			for _, ts := range list {
				if ts.TimestampId != 0 {
					t.Fatalf("expected TimestampId to be reset, got %d", ts.TimestampId)
				}
			}
		}
	}
}
