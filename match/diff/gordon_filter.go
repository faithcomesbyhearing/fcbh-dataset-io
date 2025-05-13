package diff

import (
	"github.com/sergi/go-diff/diffmatchpatch"
	"unicode"
)

// This class reads a pair slice and outputs a filtered pair slice
// It searches words for common patterns,
// And then assumes those that occur more than some threshold are false positives
// So it alters the diffs to make those differences disappear

type position struct {
}

func findWordPatterns(pairs []Pair) []diffPattern {
	var results = make([]diffPattern, 0)
	for _, pair := range pairs {
		//fmt.Println(pair.Diffs)
		var pattern diffPattern
		for _, diff := range pair.Diffs {
			for _, char := range diff.Text {
				if diff.Type != diffmatchpatch.DiffInsert &&
					unicode.IsSpace(char) &&
					!pattern.isEmpty() {
					//fmt.Println(pattern.String() + "|")
					results = append(results, pattern)
					pattern = diffPattern{}
				} else {
					pattern.appendDiff(diff.Type, char)
				}
			}
		}
		if !pattern.isEmpty() {
			//fmt.Println(pattern.String() + "|e")
			results = append(results, pattern)
		}
	}
	return results
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
