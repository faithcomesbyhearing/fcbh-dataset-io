package diff

import (
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"unicode"
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

func filter(pairs []Pair) []Pair {
	var results []Pair
	matches := findWordPatterns(pairs)
	fmt.Println("matches", len(matches))
	// find matches that exceed some limit
	// Edit matches by erasing inserts and changing delete to equal
	// Note this process is destructive meaning the counts are going to be incorrect
	return results
}

func findWordPatterns(pairs []Pair) map[string][]position {
	var results = make(map[string][]position)
	for pairIndex, pair := range pairs {
		fmt.Println(pair.Diffs)
		var pattern diffPattern
		var startDiffIndex = 0
		for diffIndex, diff := range pair.Diffs {
			for charIndex, char := range diff.Text {
				if diff.Type != diffmatchpatch.DiffInsert &&
					unicode.IsSpace(char) &&
					!pattern.isEmpty() {
					key := pattern.String()
					pos := position{pairIndex, startDiffIndex, charIndex}
					fmt.Println(key+"| -> ", pos)
					results[key] = append(results[key], pos)
					pattern = diffPattern{}
					startDiffIndex = diffIndex // +1 is correct when we are at the end of the diffItem
				} else {
					pattern.appendDiff(diff.Type, char)
				}
			}
		}
		if !pattern.isEmpty() {
			key := pattern.String()
			pos := position{pairIndex, startDiffIndex, -1}
			fmt.Println(key+"|e -> ", pos)
			results[key] = append(results[key], pos)
		}
	}
	return results
}

type position struct {
	pairIndex int
	diffIndex int
	charIndex int
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
	if diffType != d.currType || len(d.parts) == 0 {
		if diffType == diffmatchpatch.DiffInsert {
			d.parts = append(d.parts, '+')
		} else if diffType == diffmatchpatch.DiffDelete {
			d.parts = append(d.parts, '-')
		} else {
			d.parts = append(d.parts, '=')
		}
		d.currType = diffType
	}
	d.parts = append(d.parts, char)
}
