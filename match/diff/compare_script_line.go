package diff

import (
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"sort"
)

func (c *Compare) compareScriptLines(mediaType request.MediaType) ([]Pair, *log.Status) {
	var results []Pair
	baseMap, status := c.selectScriptLines(c.baseDb, true)
	if status != nil {
		return results, status
	}
	compMap, status := c.selectScriptLines(c.database, false)
	if status != nil {
		return results, status
	}
	for lineNum, pair := range compMap {
		if pair.AudioFile != "" {
			basePair := baseMap[lineNum]
			pair.Base = basePair.Base
			results = append(results, pair)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].ScriptNum < results[j].ScriptNum
	})
	results = c.cleanUpPairs(results, mediaType)
	c.isPairsLatin(results)
	for _, pair := range results {
		c.diffPair(pair)
	}
	return c.results, nil // c.results was filtered by c.diffPair
}

func (c *Compare) selectScriptLines(database db.DBAdapter, isBase bool) (map[string]Pair, *log.Status) {
	var results = make(map[string]Pair)
	query := `SELECT script_id, book_id, chapter_num, verse_str, script_num, script_text, uroman, audio_file FROM scripts`
	rows, err := database.DB.Query(query)
	if err != nil {
		return results, log.Error(c.ctx, 500, err, `Error reading rows in selectScriptLines`)
	}
	defer rows.Close()
	for rows.Next() {
		var p Pair
		var t PairText
		err = rows.Scan(&t.ScriptId, &p.Ref.BookId, &p.Ref.ChapterNum, &p.Ref.VerseStr, &p.ScriptNum,
			&t.Text, &t.Uroman, &p.AudioFile)
		if err != nil {
			return results, log.Error(c.ctx, 500, err, `Error scanning in selectScriptLines`)
		}
		if isBase {
			p.Base = t
		} else {
			p.Comp = t
		}
		results[p.ScriptNum] = p
	}
	err = rows.Err()
	if err != nil {
		return results, log.Error(c.ctx, 500, err, `Error at end of rows in selectScriptLines`)
	}
	return results, nil
}

func (c *Compare) cleanUpPairs(pairs []Pair, mediaType request.MediaType) []Pair {
	for i := range pairs {
		pairs[i].Base.Text = c.cleanup(pairs[i].Base.Text)
		pairs[i].Comp.Text = c.cleanup(pairs[i].Comp.Text)
		pairs[i].Base.Uroman = c.cleanup(pairs[i].Base.Uroman)
		pairs[i].Comp.Uroman = c.cleanup(pairs[i].Comp.Uroman)
		if mediaType == request.TextScript {
			pairs[i].Base.Text = c.verseRm.ReplaceAllString(pairs[i].Base.Text, ``)
		}
	}
	return pairs
}

func (c *Compare) isPairsLatin(pairs []Pair) {
	// This needs to be written if the user needs to always see latin
	// See SetIsLatin
	c.isLatin.Valid = true
	c.isLatin.Bool = true
}
