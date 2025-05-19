package diff

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGordonFilter(t *testing.T) {
	pairs := loadPairs(t)
	matchThreshold := 40
	newPairs := GordonFilter(pairs, matchThreshold)
	pairsCopy := loadPairs(t)
	for i := 0; i < 10; i++ {
		fmt.Println("pairs")
		for _, diff := range pairsCopy[i].Diffs {
			fmt.Print(diff, "; ")
		}
		fmt.Println()
		fmt.Println("new Pairs")
		for _, diff := range newPairs[i].Diffs {
			fmt.Print(diff, "; ")
		}
		fmt.Println("\n**************")
	}
}

func TestConvertToCharDiff(t *testing.T) {
	pairs := loadPairs(t)
	tmpPairs := convertDiffToCharDiff(pairs[:10])
	for i := 0; i < 10; i++ {
		fmt.Println(pairs[i].Diffs)
		fmt.Println()
		cDiff := tmpPairs[i].charDiffs
		for _, cdif := range cDiff {
			fmt.Print(cdif.dType, " ", string(cdif.char), "; ")
		}
		fmt.Println("\n**************")
	}
}

func TestCharDiffSymetricTest(t *testing.T) {
	pairs := loadPairs(t)
	tmpPairs := convertDiffToCharDiff(pairs)
	for i := 0; i < len(pairs); i++ {
		pairs[i].Diffs = pairs[i].Diffs[:0]
	}
	pairs2 := convertCharDiffToDiff(tmpPairs, pairs)
	for i := 0; i < len(pairs2); i++ {
		if len(pairs2[i].Diffs) != len(pairs[i].Diffs) {
			t.Error("diffs not equal")
		}
		for j := 0; j < len(pairs[i].Diffs); j++ {
			if pairs2[i].Diffs[j] != pairs[i].Diffs[j] {
				t.Error("diffs not equal")
			}
		}
	}
}

func TestFindPatterns(t *testing.T) {
	pairs := loadPairs(t)
	tmpPairs := convertDiffToCharDiff(pairs)
	results := findWordPatterns(tmpPairs[:1])
	for i := 0; i < 1; i++ {
		fmt.Println(pairs[i].Diffs)
		fmt.Println()
		chr := tmpPairs[i].charDiffs
		for _, vrs := range chr {
			fmt.Print(vrs.dType, string(vrs.char), "; ")
		}
		fmt.Println()
		for pattern, item := range results {
			fmt.Println(pattern, item)
		}
		fmt.Println("**************")
	}
	if len(results) != 241334 {
		t.Error("Got length results", len(results))
	}
}

func TestPrunePatterns(t *testing.T) {
	pairs := loadPairs(t)
	tmpPairs := convertDiffToCharDiff(pairs)
	results := findWordPatterns(tmpPairs)
	results = prunePatterns(results, 40)
}

func loadPairs(t *testing.T) []Pair {
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
	return pairs
}
