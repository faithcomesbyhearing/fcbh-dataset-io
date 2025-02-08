package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/generic"
	_ "github.com/mattn/go-sqlite3"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type AudioError struct {
	Reference string
	WordSeq   int
	Word      string
	EType     string
	ScriptId  int64
	WordId    int64
	FAError   float64
	Duration  float64
}

func main() {
	directory := "./match/research/fa_error/data"
	audioErr, err := ReadErrorCSV(filepath.Join(directory, "N2CUL_MNT_errors.csv"))
	if err != nil {
		panic(err)
	}
	//for _, ae := range audioErr {
	//	fmt.Println(ae)
	//}
	var DB *sql.DB
	DB, err = sql.Open("sqlite3", filepath.Join(directory, "N2CUL_MNT.db"))
	if err != nil {
		panic(err)
	}
	defer DB.Close()
	for i := range audioErr {
		audioErr[i], err = SelectWord(DB, audioErr[i])
		if err != nil {
			panic(err)
		}
		//fmt.Println(audioErr[i])
	}
}

func ReadErrorCSV(filename string) ([]AudioError, error) {
	var result []AudioError
	file, err := os.Open(filename)
	if err != nil {
		return result, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return result, err
	}

	var errors []AudioError
	for _, rec := range records { // Skip header row
		var ae AudioError
		ae.Reference = strings.TrimSpace(rec[0])
		ae.WordSeq, err = strconv.Atoi(strings.TrimSpace(rec[1]))
		if err != nil {
			fmt.Println(err)
		}
		ae.Word = strings.TrimSpace(rec[2])
		ae.EType = strings.TrimSpace(rec[3])
		errors = append(errors, ae)
	}
	return errors, nil
}

func SelectWord(db *sql.DB, ae AudioError) (AudioError, error) {
	ref := generic.NewVerseRef(ae.Reference)
	query := `SELECT script_id, word_id, word, fa_score, word_begin_ts, word_end_ts 
			FROM words WHERE script_id IN 
			(SELECT script_id FROM scripts WHERE book_id = ? AND chapter_num = ? AND verse_str = ?)
			AND word_seq = ? AND ttype = 'W'`
	row := db.QueryRow(query, ref.BookId, ref.ChapterNum, ref.VerseStr, ae.WordSeq)
	var faScore, beginTS, endTS float64
	var word string
	err := row.Scan(&ae.ScriptId, &ae.WordId, &word, &faScore, &beginTS, &endTS)
	if err != nil {
		//return ae, err
		fmt.Println(err)
	}
	if ae.Word != word {
		fmt.Println("Ref", ae.Reference, "Expected:", ae.Word, "Found:", word)
	}
	ae.FAError = -math.Log10(faScore)
	ae.Duration = endTS - beginTS
	return ae, nil
}
