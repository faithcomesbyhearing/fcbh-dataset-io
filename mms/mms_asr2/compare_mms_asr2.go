package mms_asr

import (
	"context"
	"database/sql"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/match/diff"
	"github.com/sergi/go-diff/diffmatchpatch"
	"regexp"
	"strings"
)

type Compare struct {
	ctx         context.Context
	user        string
	baseDataset string
	dataset     string
	baseDb      db.DBAdapter
	database    db.DBAdapter
	lang        string
	baseIdent   db.Ident
	compIdent   db.Ident
	testament   request.Testament
	settings    request.CompareSettings
	replacer    *strings.Replacer
	verseRm     *regexp.Regexp
	isLatin     sql.NullBool
	diffMatch   *diffmatchpatch.DiffMatchPatch
	results     []diff.Pair
}

func (a *Compare) compareToASRTable() ([]diff.Pair, *log.Status) {
	var results []diff.Pair
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
		var p diff.Pair
		var uRoman sql.NullString
		var text sql.NullString
		err = rows.Scan(&p.Base.ScriptId, &p.Ref.BookId, &p.Ref.ChapterNum, &p.Ref.VerseStr, &p.ScriptNum,
			&p.AudioFile, &p.BeginTS, &p.EndTS, &p.Base.Text, &p.Base.Uroman, &text, &uRoman)
		if err != nil {
			return results, log.Error(a.ctx, 500, err, query)
		}
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

func (a *Compare) diffPairs(pairs []diff.Pair) {
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

func (c *Compare) isMatch(diffs []diffmatchpatch.Diff) bool {
	for _, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffInsert || diff.Type == diffmatchpatch.DiffDelete {
			if len(strings.TrimSpace(diff.Text)) > 0 {
				return false
			}
		}
	}
	return true
}
