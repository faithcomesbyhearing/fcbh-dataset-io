package diff

import (
	"database/sql"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"strings"
)

/*
This code is in development as of Nov 12, 2025 GNG
*/

func (a *Compare) CompareASR2() ([]Pair, string, string, *log.Status) {
	pairs, status := a.selectASRPairs()
	if status != nil {
		return pairs, "", "", status
	}
	for i := range pairs {
		pairs[i].Base.Text = a.cleanup(pairs[i].Base.Text)
		pairs[i].Base.Uroman = a.cleanup(pairs[i].Base.Uroman)
		pairs[i].Comp.Text = a.cleanup(pairs[i].Comp.Text)
		pairs[i].Comp.Uroman = a.cleanup(pairs[i].Comp.Uroman)
	}
	a.diffPairs(pairs)
	fileMap, status := a.generateBookChapterFilenameMap()
	return pairs, fileMap, a.lang, status
}

func (a *Compare) selectASRPairs() ([]Pair, *log.Status) {
	var results []Pair
	query := `SELECT s.script_id, s.book_id, s.chapter_num, s.verse_str, s.script_num,
			s.audio_file, s.script_begin_ts, s.script_end_ts, s.script_text, s.uroman,
			a.script_text, a.uroman
			FROM scripts s LEFT OUTER JOIN asr a ON s.script_id = a.script_id`
	rows, err := a.database.DB.Query(query)
	if err != nil {
		return results, log.Error(a.ctx, 500, err, query)
	}
	defer rows.Close()
	for rows.Next() {
		var p Pair
		var uRoman sql.NullString
		var text sql.NullString
		err = rows.Scan(&p.Base.ScriptId, &p.Ref.BookId, &p.Ref.ChapterNum, &p.Ref.VerseStr, &p.ScriptNum,
			&p.AudioFile, &p.BeginTS, &p.EndTS, &p.Base.Text, &p.Base.Uroman, &text, &uRoman)
		if err != nil {
			return results, log.Error(a.ctx, 500, err, query)
		}
		p.ScriptNum = "" // non-blank script num will display as line num on report
		p.Comp.ScriptId = p.Base.ScriptId
		if text.Valid {
			p.Comp.Text = text.String
		} else {
			p.Comp.Text = ""
		}
		if uRoman.Valid {
			p.Comp.Uroman = uRoman.String
		} else {
			p.Comp.Uroman = ""
		}
		results = append(results, p)
	}
	err = rows.Err()
	if err != nil {
		return results, log.Error(a.ctx, 500, err, query)
	}
	return results, nil
}

func (a *Compare) diffPairs(pairs []Pair) {
	for i, p := range pairs {
		if len(p.Base.Text) > 0 || len(p.Comp.Text) > 0 {
			baseText := strings.TrimSpace(p.Base.Text)
			compText := strings.TrimSpace(p.Comp.Text)
			diffs := a.diffMatch.DiffMain(baseText, compText, false)
			aDiff := a.diffMatch.DiffCleanupMerge(diffs)
			if !a.isMatch(aDiff) {
				pairs[i].HTML = a.diffMatch.DiffPrettyHtml(aDiff)
			}
			pairs[i].Diffs = aDiff
		}
	}
}
