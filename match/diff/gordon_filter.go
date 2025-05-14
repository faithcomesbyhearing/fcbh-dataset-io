package diff

import (
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"unicode"
	"unicode/utf8"
)

/**
This class reads a pair slice and outputs a filtered pair slice
It searches words for common patterns,
And then assumes those that occur more than some threshold are false positives
So it alters the diffs to make those differences disappear
*/

/*
Single char solution
1) type charDiff (diff operation, char rune, remove bool)
2) func - convert to char level type
	a) iterate over pairs
	b) iterate over diff
	c) iterate over char
	d) output charDiff
3) func - findWord patterns
	a) iterate over pairs
	b) iterate over chars
	c) build patters like a+b+c=d=e=f-f
	d) breaking on equal or delete space
	e) position type (pairIndex int, charIndex int)
4) func pruneWordPattens
	a) iterate over patterns
	b) drop patterns that have less that some threshold
5) func - modifyCharDiff
	a) iterate over patterns
	b) lookup pair by pairIndex
	c) lookup char by charIndex
	d) length = len(pattern) / 2
	e) from charPos to charPos + length
	f) change delete to equal
	g) change insert to remove=true
6) func - rebuildPairs
	a) iterate over pairs
	b) iterate over char
	c) if remove = true, skip
	d) else if type is the same, append char
	e) else append entire type, start new type, append char
*/

const MatchThreshold = 50

type tmpPair struct {
	charDiffs []charDiff
}
type charDiff struct {
	dType  diffmatchpatch.Operation
	char   rune
	remove bool
}
type position struct {
	verseIndex int
	charIndex  int
}

func filter(pairs []Pair) []Pair {
	var results []Pair
	tmpPairs := convertDiffToCharDiff(pairs)
	matches := findWordPatterns(tmpPairs)
	fmt.Println("matches", len(matches))
	matches = prunePatterns(matches)
	fmt.Println("pruned matches", len(matches))
	tmpPairs = removeCommonPatterns(matches, tmpPairs)
	results = convertCharDiffToDiff(tmpPairs, pairs)
	return results
}

func convertDiffToCharDiff(pairs []Pair) []tmpPair {
	var results []tmpPair
	for _, pair := range pairs {
		var verse []charDiff
		for _, diff := range pair.Diffs {
			for _, char := range diff.Text {
				chrDiff := charDiff{dType: diff.Type, char: char, remove: false}
				verse = append(verse, chrDiff)
			}
		}
		tPair := tmpPair{charDiffs: verse}
		results = append(results, tPair)
	}
	return results
}

func convertCharDiffToDiff(tmpPairs []tmpPair, pairs []Pair) []Pair {
	if len(tmpPairs) != len(pairs) {
		fmt.Println("convertCharDiffToDiff: number of tmpPairs", len(tmpPairs), "and pairs", len(pairs))
		panic("convertCharDiffToDiff")
	}
	for i, cDiff := range tmpPairs {
		var diffs []diffmatchpatch.Diff
		var diff diffmatchpatch.Diff
		var str []rune
		var currType = cDiff.charDiffs[0].dType
		for _, vs := range cDiff.charDiffs {
			if vs.dType != currType && len(str) > 0 {
				currType = vs.dType
				diff.Text = string(str)
				diffs = append(diffs, diff)
				diff = diffmatchpatch.Diff{}
				str = str[:0] // erase str
			}
			if !vs.remove {
				diff.Type = vs.dType
				str = append(str, vs.char)
			}
		}
		pairs[i].Diffs = diffs
	}
	return pairs
}

func findWordPatterns(tmpPairs []tmpPair) map[string][]position {
	var results = make(map[string][]position)
	for pairIndex, tPair := range tmpPairs {
		//fmt.Println(verse.verses)
		var pattern diffPattern
		var startDiffIndex = 0
		for diffIndex, diff := range tPair.charDiffs {
			if diff.dType != diffmatchpatch.DiffInsert &&
				//if diff.dType == diffmatchpatch.DiffEqual &&
				unicode.IsSpace(diff.char) &&
				!pattern.isEmpty() {
				key := pattern.String()
				pos := position{pairIndex, startDiffIndex}
				//fmt.Println(key+"| -> ", pos)
				results[key] = append(results[key], pos)
				pattern = diffPattern{}
				startDiffIndex = diffIndex + 1
			} else {
				pattern.appendDiff(diff.dType, diff.char)
			}
		}
		if !pattern.isEmpty() {
			key := pattern.String()
			pos := position{pairIndex, startDiffIndex}
			//fmt.Println(key+"|e -> ", pos)
			results[key] = append(results[key], pos)
		}
	}
	return results
}

func prunePatterns(matches map[string][]position) map[string][]position {
	var results = make(map[string][]position)
	fmt.Println("before", len(matches), "matches")
	for pattern, pos := range matches {
		if len(pos) > MatchThreshold {
			results[pattern] = pos
		}
	}
	fmt.Printf("final counts: %d\n", len(results))
	return results
}

func removeCommonPatterns(matches map[string][]position, tmpPairs []tmpPair) []tmpPair {
	fmt.Println("remove", len(matches), "matches")
	for pattern, poses := range matches {
		fmt.Println("removing", pattern, "num", len(poses))
		for _, pos := range poses {
			verse := tmpPairs[pos.verseIndex]
			numChars := utf8.RuneCountInString(pattern) / 2
			//fmt.Println(pattern, "numChars", numChars, "charIndex", pos.charIndex, "len", len(verse.charDiffs))
			for i := pos.charIndex; i < pos.charIndex+numChars; i++ {
				//fmt.Println("Fix", string(verse.charDiffs[i].char))
				if verse.charDiffs[i].dType == diffmatchpatch.DiffDelete {
					verse.charDiffs[i].dType = diffmatchpatch.DiffEqual
				} else if verse.charDiffs[i].dType == diffmatchpatch.DiffInsert {
					verse.charDiffs[i].remove = true
				}
			}
			tmpPairs[pos.verseIndex] = verse
		}
	}
	return tmpPairs
}

type diffPattern struct {
	currType diffmatchpatch.Operation
	parts    []rune
}

func (d *diffPattern) String() string {
	return string(d.parts)
}

func (d *diffPattern) isEmpty() bool {
	return len(d.parts) == 0
}

func (d *diffPattern) appendDiff(diffType diffmatchpatch.Operation, char rune) {
	//if diffType != d.currType || len(d.parts) == 0 {
	if diffType == diffmatchpatch.DiffInsert {
		d.parts = append(d.parts, '+')
	} else if diffType == diffmatchpatch.DiffDelete {
		d.parts = append(d.parts, '-')
	} else {
		d.parts = append(d.parts, '=')
	}
	//d.currType = diffType
	//}
	d.parts = append(d.parts, char)
}
