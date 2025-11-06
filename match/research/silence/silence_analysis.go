package main

import (
	"context"
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

func main() {
	ctx := context.Background()
	log.SetOutput("stderr")
	user := request.GetTestUser()
	conn, status := db.NewerDBAdapter(ctx, false, user, "N2MZJSIM")
	if status != nil {
		os.Exit(1)
	}
	type SilenceRec struct {
		ScriptId int64
		WordId   int64
		BeginTS  float64
		EndTS    float64
		Duration float64
		Silence  float64
	}
	var query = `SELECT w.script_id, w.word_id, w.word_begin_ts, w.word_end_ts, s.script_end_ts
		FROM scripts s JOIN words w ON s.script_id = w.script_id
		WHERE w.ttype = 'W'`
	rows, err := conn.DB.Query(query)
	if err != nil {
		status = log.Error(ctx, 500, err, "Error in SQL query of silence")
		os.Exit(1)
	}
	defer rows.Close()
	var results []SilenceRec
	var scriptEndTS float64
	var priorEndTS float64
	var priorScriptEndTS float64
	for rows.Next() {
		var s SilenceRec
		err = rows.Scan(&s.ScriptId, &s.WordId, &s.BeginTS, &s.EndTS, &scriptEndTS)
		if err != nil {
			status = log.Error(ctx, 500, err, `Error scanning in Select Silence`)
			os.Exit(1)
		}
		if len(results) == 0 {
			s.Silence = s.BeginTS
		} else {
			s.Silence = s.BeginTS - priorEndTS
			if s.Silence < 0 {
				s.Silence = priorScriptEndTS - priorEndTS + s.BeginTS
			}
		}
		s.Duration = s.EndTS - s.BeginTS
		priorEndTS = s.EndTS
		priorScriptEndTS = scriptEndTS
		results = append(results, s)
	}
	err = rows.Err()
	if err != nil {
		status = log.Error(ctx, 500, err, `Error at end of rows in Silence`)
		os.Exit(1)
	}
	var silences []float64
	for _, result := range results {
		silences = append(silences, result.Silence)
		if result.Silence < 0.0 {
			fmt.Println(result.Silence, result.ScriptId, result.WordId)
		}
	}
	compute(silences, "silence", "N2MZJSIM", "")
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

// select script_id, word_id, timestamps
// compute duration and silence
// silence of the first is begin of zeroth
// copy in compute

/*

	var chars []generic.AlignChar
	var query = `SELECT s.audio_file, s.script_id, s.book_id, s.chapter_num, s.verse_str,
				w.word_id, w.word, c.char_id, c.seq, c.uroman, c.start_ts, c.end_ts, c.fa_score
				FROM scripts s JOIN words w ON s.script_id = w.script_id
				JOIN chars c ON w.word_id = c.word_id
				WHERE w.ttype = 'W'
				ORDER BY c.char_id`
	rows, err := d.DB.Query(query)
*/
