package diff

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/sergi/go-diff/diffmatchpatch"
	"strings"
	"unicode"
	"unicode/utf8"
)

/**
This file reads a pair slice and outputs a filtered pair slice
It searches words for common patterns,
And then assumes those that occur more than some threshold are false positives
So it alters the diffs to make those differences disappear
*/

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

func GordonFilter(ctx context.Context, pairs []Pair, user string, dbPath string, matchThreshold int) ([]Pair, *log.Status) {
	var results []Pair
	tmpPairs := convertDiffToCharDiff(pairs)
	wordMap, status := selectWords(ctx, user, dbPath)
	if status != nil {
		return results, status
	}
	//matches := findWordPatterns(tmpPairs)
	matches := findDiscrepancyPatterns(tmpPairs, wordMap)
	fmt.Println("matches", len(matches))
	matches = prunePatterns(matches, matchThreshold)
	fmt.Println("pruned matches", len(matches))
	tmpPairs = removeCommonPatterns(matches, tmpPairs)
	results = convertCharDiffToDiff(tmpPairs, pairs)
	diffMatch := diffmatchpatch.New()
	for i := range results {
		results[i].HTML = diffMatch.DiffPrettyHtml(results[i].Diffs)
	}
	return results, nil
}

func selectWords(ctx context.Context, user string, dbPath string) (map[string]bool, *log.Status) {
	var results = make(map[string]bool)
	conn, status := db.NewerDBAdapter(ctx, false, user, dbPath)
	if status != nil {
		return results, status
	}
	var query = `SELECT distinct word FROM words WHERE ttype = 'W'`
	rows, err := conn.DB.Query(query)
	if err != nil {
		return results, log.Error(ctx, 500, err, "Error during gordon filter Select Words.")
	}
	defer rows.Close()
	for rows.Next() {
		var word string
		err = rows.Scan(&word)
		if err != nil {
			return results, log.Error(ctx, 500, err, "Error during gordon filter Select Words.")
		}
		word = strings.ToLower(word)
		minusWord := "-" + strings.Join(strings.Split(word, ""), "-")
		results[minusWord] = true
		plusWord := "+" + strings.Join(strings.Split(word, ""), "+")
		results[plusWord] = true
	}
	err = rows.Err()
	if err != nil {
		log.Warn(ctx, err, query)
	}
	conn.Close()
	return results, nil
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
		var cleanDiff = make([]charDiff, 0, len(cDiff.charDiffs))
		for _, diff := range cDiff.charDiffs {
			if !diff.remove {
				cleanDiff = append(cleanDiff, diff)
			}
		}
		var diffs []diffmatchpatch.Diff
		var str []rune
		var currType = cleanDiff[0].dType
		for _, vs := range cleanDiff {
			if currType != vs.dType {
				var diff diffmatchpatch.Diff
				diff.Type = currType
				diff.Text = string(str)
				diffs = append(diffs, diff)
				str = str[:0] // erase str
				currType = vs.dType
			}
			str = append(str, vs.char)
		}
		var diff diffmatchpatch.Diff
		diff.Type = currType
		diff.Text = string(str)
		diffs = append(diffs, diff)
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

func findDiscrepancyPatterns(tmpPairs []tmpPair, wordMap map[string]bool) map[string][]position {
	var results = make(map[string][]position)
	for pairIndex, tPair := range tmpPairs {
		var pattern diffPattern
		var startDiffIndex int
		for diffIndex, diff := range tPair.charDiffs {
			if diff.dType == diffmatchpatch.DiffEqual {
				if !pattern.isEmpty() {
					key := pattern.String()
					_, isWord := wordMap[key]
					if !isWord {
						pos := position{pairIndex, startDiffIndex}
						results[key] = append(results[key], pos)
					}
					pattern = diffPattern{}
				}
			} else {
				if pattern.isEmpty() {
					startDiffIndex = diffIndex
				}
				pattern.appendDiff(diff.dType, diff.char)
			}
		}
		if !pattern.isEmpty() {
			key := pattern.String()
			_, isWord := wordMap[key]
			if !isWord {
				pos := position{pairIndex, startDiffIndex}
				results[key] = append(results[key], pos)
			}
		}
	}
	return results
}

func prunePatterns(matches map[string][]position, matchThreshold int) map[string][]position {
	var results = make(map[string][]position)
	fmt.Println("before", len(matches), "matches")
	for pattern, pos := range matches {
		if len(pos) > matchThreshold {
			results[pattern] = pos
		}
	}
	fmt.Printf("final counts: %d\n", len(results))
	return results
}

func removeCommonPatterns(matches map[string][]position, tmpPairs []tmpPair) []tmpPair {
	fmt.Println("remove", len(matches), "matches")
	for pattern, poses := range matches {
		//fmt.Println("removing", pattern, "num", len(poses))
		for _, pos := range poses {
			verse := tmpPairs[pos.verseIndex]
			numChars := utf8.RuneCountInString(pattern) / 2
			lastChar := pos.charIndex + numChars
			if lastChar < len(verse.charDiffs) {
				lastChar++
			}
			//fmt.Println(pattern, "numChars", numChars, "charIndex", pos.charIndex, "len", len(verse.charDiffs))
			for i := pos.charIndex; i < lastChar; i++ {
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
	if diffType == diffmatchpatch.DiffInsert {
		d.parts = append(d.parts, '+')
	} else if diffType == diffmatchpatch.DiffDelete {
		d.parts = append(d.parts, '-')
	} else {
		d.parts = append(d.parts, '=')
	}
	d.parts = append(d.parts, char)
}
