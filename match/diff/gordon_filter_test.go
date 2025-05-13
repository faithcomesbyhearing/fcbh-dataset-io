package diff

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFindPatterns(t *testing.T) {
	filePath := filepath.Join(os.Getenv("HOME"), "Downloads", "N2ANLBSM_audio_compare.json")
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	var pairs []Pair
	err = json.Unmarshal(bytes, &pairs)
	if err != nil {
		t.Fatal(err)
	}
	results := findWordPatterns(pairs[0:10])
	if len(results) != 241334 {
		t.Error("Got length results", len(results))
	}
}
