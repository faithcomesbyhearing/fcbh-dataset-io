package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/generic"
	_ "github.com/mattn/go-sqlite3"
	"gonum.org/v1/gonum/stat"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type AudioError struct {
	Reference string  `json:"reference"`
	WordSeq   int     `json:"word_seq"`
	Word      string  `json:"word"`
	EType     string  `json:"type"`
	ScriptId  int64   `json:"script_id"`
	WordId    int64   `json:"word_id"`
	FAError   float64 `json:"fa_error"`
	Duration  float64 `json:"duration"`
}

func main() {
	directory := "./match/research/fa_error/data"
	audioErr, err := readErrorCSV(filepath.Join(directory, "N2CUL_MNT_errors.csv"))
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
		audioErr[i], err = selectWord(DB, audioErr[i])
		if err != nil {
			panic(err)
		}
		//fmt.Println(audioErr[i])
	}
	content, err := json.MarshalIndent(audioErr, "", "    ")
	if err != nil {
		panic(err)
	}
	_, err = os.Stdout.Write(content)
	if err != nil {
		panic(err)
	}
	var floats []float64
	for _, ae := range audioErr {
		floats = append(floats, float64(ae.FAError))
	}
	compute(floats, "Words", "N2CUL_MNT", directory)
	var allScores []float64
	allScores, err = selectAllWordScores(DB)
	if err != nil {
		panic(err)
	}
	compute(allScores, "All Words", "N2CUL_MNT", directory)
}

func readErrorCSV(filename string) ([]AudioError, error) {
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

func selectWord(db *sql.DB, ae AudioError) (AudioError, error) {
	type word struct {
		scriptId int64
		wordId   int64
		word     string
		faScore  float64
		beginTS  float64
		endTS    float64
	}
	ref := generic.NewVerseRef(ae.Reference)
	query := `SELECT script_id, word_id, word, fa_score, word_begin_ts, word_end_ts 
			FROM words WHERE ttype = 'W' AND script_id IN 
			(SELECT script_id FROM scripts WHERE book_id = ? AND chapter_num = ? AND verse_str = ?)`
	row, err := db.Query(query, ref.BookId, ref.ChapterNum, ref.VerseStr)
	if err != nil {
		return ae, err
	}
	var words []word
	for row.Next() {
		var w word
		err = row.Scan(&w.scriptId, &w.wordId, &w.word, &w.faScore, &w.beginTS, &w.endTS)
		if err != nil {
			return ae, err
		}
		words = append(words, w)
	}
	found := words[ae.WordSeq-1]
	if ae.Word != found.word {
		fmt.Println("Ref", ae.Reference, "Expected:", ae.Word, "Found:", found.word)
	} else {
		ae.ScriptId = found.scriptId
		ae.WordId = found.wordId
		ae.FAError = -math.Log10(found.faScore)
		ae.Duration = found.endTS - found.beginTS
	}
	return ae, nil
}

func selectAllWordScores(db *sql.DB) ([]float64, error) {
	var result []float64
	query := `SELECT fa_score FROM words WHERE ttype = 'W'`
	row, err := db.Query(query)
	if err != nil {
		return result, err
	}
	for row.Next() {
		var score float64
		err = row.Scan(&score)
		if err != nil {
			return result, err
		}
		if score > 0.0 {
			result = append(result, -math.Log10(score))
		}
	}
	return result, nil
}

func compute(floats []float64, desc string, stockNo string, directory string) {
	var filename = filepath.Join(directory, stockNo+"_"+desc+".txt")
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	mean := stat.Mean(floats, nil)
	write(file, "Mean: ", strconv.FormatFloat(mean, 'f', 2, 64))
	stdDev := stat.StdDev(floats, nil)
	write(file, "StdDev: ", strconv.FormatFloat(stdDev, 'f', 2, 64))

	var mini = math.Inf(1)
	var maxi = 0.0
	for _, er := range floats {
		if er < mini {
			mini = er
		}
		if er > maxi {
			maxi = er
		}
	}
	write(file, "Minimum: ", strconv.FormatFloat(mini, 'f', 2, 64))
	write(file, "Maximum: ", strconv.FormatFloat(maxi, 'f', 2, 64))
	// Skewness (asymmetry of distribution)
	skewness := stat.Skew(floats, nil)
	write(file, "Skewness: ", strconv.FormatFloat(skewness, 'f', 2, 64))
	// Kurtosis (shape of distribution)
	kurtosis := stat.ExKurtosis(floats, nil)
	write(file, "Kurtosis: ", strconv.FormatFloat(kurtosis, 'f', 2, 64))
	// Percentile
	write(file, "\nPercentiles")
	sort.Float64s(floats)
	for _, percent := range []float64{
		0.00, 0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.07, 0.08, 0.09,
		0.10, 0.11, 0.12, 0.13, 0.14, 0.15, 0.16, 0.17, 0.18, 0.19,
		0.20, 0.21, 0.22, 0.23, 0.24, 0.25, 0.26, 0.27, 0.28, 0.29,
		0.30, 0.31, 0.32, 0.33, 0.34, 0.35, 0.36, 0.37, 0.38, 0.39,
		0.40, 0.41, 0.42, 0.43, 0.44, 0.45, 0.46, 0.47, 0.48, 0.49,
		0.50, 0.51, 0.52, 0.53, 0.54, 0.55, 0.56, 0.57, 0.58, 0.59,
		0.60, 0.61, 0.62, 0.63, 0.64, 0.65, 0.66, 0.67, 0.68, 0.69,
		0.70, 0.71, 0.72, 0.73, 0.74, 0.75, 0.76, 0.77, 0.78, 0.79,
		0.80, 0.81, 0.82, 0.83, 0.84, 0.85, 0.86, 0.87, 0.88, 0.89,
		0.90, 0.91, 0.92, 0.93, 0.94, 0.95, 0.96, 0.97, 0.98, 0.99,
		0.991, 0.992, 0.993, 0.994, 0.995, 0.996, 0.997, 0.998, 0.999,
		0.9991, 0.9992, 0.9993, 0.9994, 0.9995, 0.9996, 0.9997, 0.9998, 0.99999,
	} {
		percentile := stat.Quantile(percent, stat.Empirical, floats, nil)
		percentStr := strconv.FormatFloat((percent * 100.0), 'f', 2, 64)
		write(file, "Percentile ", percentStr, ": ", strconv.FormatFloat(percentile, 'f', 2, 64))
	}
	// Histogram
	write(file, "\nHISTOGRAM")
	var histogram = make(map[int]int)
	for _, er := range floats {
		histogram[int(er)]++
	}
	var keys []int
	for k := range histogram {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	numFAError := len(floats)
	for _, cat := range keys {
		pct := float64(histogram[cat]) / float64(numFAError) * 100.0
		write(file, "Cat: ", strconv.Itoa(cat), "-", strconv.Itoa(cat+1), " = ", strconv.FormatFloat(pct, 'f', 4, 64), "%")
	}
}

func write(file *os.File, args ...string) {
	for _, arg := range args {
		_, _ = file.WriteString(arg)
	}
	_, _ = file.WriteString("\n")
}
