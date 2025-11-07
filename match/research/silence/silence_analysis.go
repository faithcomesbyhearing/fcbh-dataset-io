package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"gonum.org/v1/gonum/stat"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type SilenceRec struct {
	ScriptId    int64
	BookId      string
	ChapterNum  int
	VerseStr    string
	ScriptEndTS float64
	WordId      int64
	WordSeq     int
	BeginTS     float64
	EndTS       float64
	Duration    float64
	Silence     float64
}

func (s *SilenceRec) Ref() string {
	return s.BookId + " " + strconv.Itoa(s.ChapterNum) + ":" + s.VerseStr + "." + strconv.Itoa(s.WordSeq)
}
func (s *SilenceRec) BeginT() string {
	return strconv.FormatFloat(s.BeginTS, 'f', 2, 64)
}
func (s *SilenceRec) EndT() string {
	return strconv.FormatFloat(s.EndTS, 'f', 2, 64)
}
func (s *SilenceRec) EndScript() string {
	return strconv.FormatFloat(s.ScriptEndTS, 'f', 2, 64)
}
func (s *SilenceRec) Dur() string {
	return strconv.FormatFloat(s.Duration, 'f', 2, 64)
}
func (s *SilenceRec) Silen() string {
	return strconv.FormatFloat(s.Silence, 'f', 2, 64)
}

func main() {
	ctx := context.Background()
	log.SetOutput("stderr")
	user := request.GetTestUser()
	conn, status := db.NewerDBAdapter(ctx, false, user, "N2MZJSIM")
	if status != nil {
		os.Exit(1)
	}
	results := selectSilences(ctx, nil, conn)
	var silences []float64
	for _, result := range results {
		silences = append(silences, result.Silence)
		if result.Silence < 0.0 {
			fmt.Println(result.Silence, result.ScriptId, result.WordId)
		}
	}
	compute(silences, "silence", "N2MZJSIM", "./match/research/silence")
	//findMaxSilenceInRefs(ctx, conn)
	findWorstVerses(results)
	//dumpTimestamps(results)
}

func compute(floats []float64, desc string, stockNo string, directory string) {
	var filename = filepath.Join(directory, stockNo+"_"+desc+".txt")
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	log.Info(context.Background(), "Output filename", filename)
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

type ref struct {
	bookId     string
	chapterNum int
	verseStr   string
}

func findMaxSilenceInRefs(ctx context.Context, conn db.DBAdapter) {
	var refs []ref
	refs = append(refs, ref{bookId: "MAT", chapterNum: 5, verseStr: "36"})
	refs = append(refs, ref{bookId: "MAT", chapterNum: 8, verseStr: "9"})
	refs = append(refs, ref{bookId: "MAT", chapterNum: 12, verseStr: "33"})
	refs = append(refs, ref{bookId: "MAT", chapterNum: 12, verseStr: "37"})
	refs = append(refs, ref{bookId: "MAT", chapterNum: 17, verseStr: "22"}) //"22-23"})
	refs = append(refs, ref{bookId: "MAT", chapterNum: 22, verseStr: "32"})
	refs = append(refs, ref{bookId: "MAT", chapterNum: 25, verseStr: "38"})
	refs = append(refs, ref{bookId: "MAT", chapterNum: 26, verseStr: "6"}) //"6-7"})
	refs = append(refs, ref{bookId: "MRK", chapterNum: 4, verseStr: "5"})
	refs = append(refs, ref{bookId: "MRK", chapterNum: 4, verseStr: "24"})
	refs = append(refs, ref{bookId: "MRK", chapterNum: 9, verseStr: "39"})
	refs = append(refs, ref{bookId: "MRK", chapterNum: 10, verseStr: "38"})
	refs = append(refs, ref{bookId: "LUK", chapterNum: 4, verseStr: "19"})
	refs = append(refs, ref{bookId: "LUK", chapterNum: 11, verseStr: "42"})
	refs = append(refs, ref{bookId: "LUK", chapterNum: 20, verseStr: "38"})
	refs = append(refs, ref{bookId: "LUK", chapterNum: 23, verseStr: "29"})
	refs = append(refs, ref{bookId: "JHN", chapterNum: 3, verseStr: "21"})
	refs = append(refs, ref{bookId: "JHN", chapterNum: 5, verseStr: "22"}) //"22-23"})
	refs = append(refs, ref{bookId: "JHN", chapterNum: 19, verseStr: "6"})
	refs = append(refs, ref{bookId: "JHN", chapterNum: 21, verseStr: "19"})
	refs = append(refs, ref{bookId: "ACT", chapterNum: 8, verseStr: "39"})
	refs = append(refs, ref{bookId: "ACT", chapterNum: 12, verseStr: "10"})
	refs = append(refs, ref{bookId: "EPH", chapterNum: 3, verseStr: "6"})
	for _, r := range refs {
		silence := findMaxSilence(ctx, &r, conn)
		log.Info(ctx, r.bookId, r.chapterNum, r.verseStr, silence)
	}
}

func findMaxSilence(ctx context.Context, ref *ref, conn db.DBAdapter) float64 {
	results := selectSilences(ctx, ref, conn)
	log.Info(ctx, "---")
	log.Info(ctx, ref.bookId, ref.chapterNum, ref.verseStr)
	for _, res := range results {
		log.Info(ctx, "res", res.Silence)
	}
	var maxi = 0.0
	for i, item := range results {
		if i > 0 && item.Silence > maxi { // first one must is incorrect.
			maxi = item.Silence
		}
	}
	log.Info(ctx, "maxi", maxi)
	return maxi
}

func selectSilences(ctx context.Context, ref *ref, conn db.DBAdapter) []SilenceRec {
	var query = `SELECT w.script_id, s.book_id, s.chapter_num, s.verse_str, s.script_end_ts,
		w.word_id, w.word_seq, w.word_begin_ts, w.word_end_ts
		FROM scripts s JOIN words w ON s.script_id = w.script_id
		WHERE w.ttype = 'W'`
	var rows *sql.Rows
	var err error
	if ref == nil {
		query += " ORDER BY w.word_id"
		rows, err = conn.DB.Query(query)
	} else {
		query += " AND s.book_id = ? AND s.chapter_num = ? AND s.verse_str = ? ORDER BY w.word_id"
		rows, err = conn.DB.Query(query, ref.bookId, ref.chapterNum, ref.verseStr)
	}
	if err != nil {
		_ = log.Error(ctx, 500, err, "Error in SQL query of silence")
		os.Exit(1)
	}
	defer rows.Close()
	var results []SilenceRec
	for rows.Next() {
		var s SilenceRec
		err = rows.Scan(&s.ScriptId, &s.BookId, &s.ChapterNum, &s.VerseStr, &s.ScriptEndTS,
			&s.WordId, &s.WordSeq, &s.BeginTS, &s.EndTS)
		if err != nil {
			_ = log.Error(ctx, 500, err, `Error scanning in Select Silence`)
			os.Exit(1)
		}
		s.Duration = s.EndTS - s.BeginTS
		results = append(results, s)
	}
	err = rows.Err()
	if err != nil {
		_ = log.Error(ctx, 500, err, `Error at end of rows in Silence`)
		os.Exit(1)
	}
	results[0].Silence = results[0].BeginTS
	for i := 1; i < len(results)-1; i++ {
		if results[i].BookId != results[i+1].BookId || results[i].ChapterNum != results[i+1].ChapterNum {
			results[i].Silence = results[i].ScriptEndTS - results[i-1].BeginTS
		} else {
			results[i].Silence = results[i].BeginTS - results[i-1].EndTS
		}
	}
	last := len(results) - 1
	results[last].Silence = results[last].ScriptEndTS - results[last-1].BeginTS
	return results
}

func findWorstVerses(silences []SilenceRec) {
	sort.Slice(silences, func(i, j int) bool {
		return silences[i].Silence > silences[j].Silence
	})
	silences = silences[0:1000]
	sort.Slice(silences, func(i, j int) bool {
		return silences[i].WordId < silences[j].WordId
	})
	for _, s := range silences {
		fmt.Println(s.Ref(), s.ScriptId, s.WordId, s.Silence)
	}
}

func dumpTimestamps(silences []SilenceRec) {
	//for _, s := range silences {
	for i := 0; i < 5000; i++ {
		s := silences[i]
		fmt.Println(s.Ref(), s.BeginT(), s.EndT(), s.EndScript(), s.Dur(), s.Silen())
	}
}
