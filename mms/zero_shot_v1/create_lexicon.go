package zero_shot_v1

import (
	"context"
	"database/sql"
	"fmt"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const TokenFile = "tokens.txt"
const LexiconFile = "lexicon.txt"
const ScriptFile = "script.txt"

type text struct {
	scriptID int64
	word     string
}

func createLexiconFile(ctx context.Context, database *sql.DB) (string, *log.Status) {
	var directory string
	var err error
	if testing.Testing() {
		directory = "data"
	} else {
		directory, err = os.MkdirTemp("", "kenlm_*")
		if err != nil {
			return directory, log.Error(ctx, 500, err, "Error creating temporary directory")
		}
	}
	var tokenFile, lexiconFile, scriptFile string
	words, status := selectWords(ctx, database)
	if status != nil {
		return directory, status
	}
	tokenFile, status = createTokens(ctx, words, directory)
	if status != nil {
		return directory, status
	}
	fmt.Println("created ", tokenFile)
	lexiconFile, status = createLexicon(ctx, words, directory)
	if status != nil {
		return directory, status
	}
	fmt.Println("created ", lexiconFile)
	scriptFile, status = createScript(ctx, words, directory)
	fmt.Println("created ", scriptFile)
	return directory, status
}

func selectWords(ctx context.Context, db *sql.DB) ([]text, *log.Status) {
	var results []text
	var query = `SELECT script_id, word FROM words WHERE ttype = 'W' ORDER BY script_id, word_id`
	rows, err := db.Query(query)
	if err != nil {
		return results, log.Error(ctx, 500, err, "Error during Select Words.")
	}
	defer rows.Close()
	for rows.Next() {
		var rec text
		err = rows.Scan(&rec.scriptID, &rec.word)
		if err != nil {
			return results, log.Error(ctx, 500, err, "Error during Select Words.")
		}
		results = append(results, rec)
	}
	err = rows.Err()
	if err != nil {
		log.Warn(ctx, err, query)
	}
	return results, nil
}

// createTokens extracts unique characters from the words and writes them to tokens.txt
func createTokens(ctx context.Context, words []text, directory string) (string, *log.Status) {
	// Extract unique characters
	charSet := make(map[rune]bool)
	for _, wd := range words {
		for _, ch := range strings.ToLower(wd.word) {
			charSet[ch] = true
		}
	}
	// Convert to sorted slice
	var chars []string
	for ch := range charSet {
		chars = append(chars, string(ch))
	}
	sort.Strings(chars)
	// Write to file
	filename := filepath.Join(directory, TokenFile)
	file, err := os.Create(filename)
	if err != nil {
		return "", log.Error(ctx, 500, err, "Error during CreateTokens.")
	}
	defer file.Close()
	// Write tokens
	_, _ = file.WriteString("|\n")
	for _, ch := range chars {
		_, _ = file.WriteString(ch + "\n")
	}
	_, _ = file.WriteString("<1>\n")
	_, _ = file.WriteString("#\n")
	return filename, nil
}

// createLexicon generates a lexicon file from word list
func createLexicon(ctx context.Context, words []text, directory string) (string, *log.Status) {
	// Extract unique words
	wordSet := make(map[string]bool)
	for _, wd := range words {
		wordSet[strings.ToLower(wd.word)] = true
	}
	// Convert to sorted slice
	var uniqueWords []string
	for word := range wordSet {
		uniqueWords = append(uniqueWords, word)
	}
	sort.Strings(uniqueWords)
	// Write to file
	filename := filepath.Join(directory, LexiconFile)
	file, err := os.Create(filename)
	if err != nil {
		return "", log.Error(ctx, 500, err, "Error during Create Lexicon.")
	}
	defer file.Close()
	// Write lexicon entries
	for _, word := range uniqueWords {
		_, _ = file.WriteString(word + " ")
		for _, ch := range word {
			if ch == '-' {
				_, _ = file.WriteString("| ")
			} else {
				_, _ = file.WriteString(string(ch) + " ")
			}
		}
		_, _ = file.WriteString("|\n")
	}
	return filename, nil
}

// createText writes all words to text.txt, grouped by script_id
func createScript(ctx context.Context, words []text, directory string) (string, *log.Status) {
	if len(words) == 0 {
		return "", log.ErrorNoErr(ctx, 500, "No data to process in create script")
	}
	// Write to file
	filename := filepath.Join(directory, ScriptFile)
	file, err := os.Create(filename)
	if err != nil {
		return "", log.Error(ctx, 500, err, "Error creating script file")
	}
	defer file.Close()
	first := true
	currScriptID := words[0].scriptID
	for _, wd := range words {
		if wd.scriptID != currScriptID {
			_, _ = file.WriteString("\n")
			currScriptID = wd.scriptID
		} else if !first {
			_, _ = file.WriteString(" ")
		}
		_, _ = file.WriteString(wd.word)
		first = false
	}
	_, _ = file.WriteString("\n")
	return filename, nil
}
