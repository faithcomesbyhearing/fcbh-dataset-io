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
					startDiffIndex = diffIndex // +1 is correct when we are at the end of the diffitem
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
