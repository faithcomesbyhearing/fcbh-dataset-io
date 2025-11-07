package adapter

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"os"
	"sort"
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
	LastWord    bool
	Duration    float64
	Silence     float64
}

func SilencePruner(ctx context.Context, threshold int, conn db.DBAdapter) []int64 {
	silences := findSilence(ctx, threshold, conn)
	var scriptIds []int64
	for _, s := range silences {
		scriptIds = append(scriptIds, s.ScriptId)
	}
	return scriptIds
}

func findSilence(ctx context.Context, threshold int, conn db.DBAdapter) []SilenceRec {
	silences := selectSilences(ctx, conn)
	duplicateOnLastWord(silences)
	silences = summarizeMaxByVerse(silences)
	sort.Slice(silences, func(i, j int) bool { // descending sort
		return silences[i].Silence > silences[j].Silence
	})
	return silences[0:threshold]
}

func selectSilences(ctx context.Context, conn db.DBAdapter) []SilenceRec {
	var query = `SELECT w.script_id, s.book_id, s.chapter_num, s.verse_str, s.script_end_ts,
		w.word_id, w.word_seq, w.word_begin_ts, w.word_end_ts
		FROM scripts s JOIN words w ON s.script_id = w.script_id
		WHERE w.ttype = 'W' ORDER BY w.word_id`
	rows, err := conn.DB.Query(query)
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
			results[i].Silence = results[i].ScriptEndTS - results[i].EndTS
		} else {
			results[i].Silence = results[i+1].BeginTS - results[i].EndTS
		}
	}
	last := len(results) - 1
	results[last].Silence = results[last].ScriptEndTS - results[last].EndTS
	return results
}

func summarizeMaxByVerse(records []SilenceRec) []SilenceRec {
	if len(records) == 0 {
		return nil
	}
	maxSilenceMap := make(map[int64]*SilenceRec)
	for i := range records {
		rec := &records[i]
		if existing, found := maxSilenceMap[rec.ScriptId]; !found || rec.Silence > existing.Silence {
			maxSilenceMap[rec.ScriptId] = rec
		}
	}
	result := make([]SilenceRec, 0, len(maxSilenceMap))
	for _, rec := range maxSilenceMap {
		result = append(result, *rec)
	}
	return result
}

func duplicateOnLastWord(records []SilenceRec) {
	for i := 0; i < len(records)-1; i++ {
		if records[i].ScriptId != records[i+1].ScriptId {
			records[i].LastWord = true
			if records[i+1].Silence < records[i].Silence {
				records[i+1].Silence = records[i].Silence
			}
		}
	}
}
