package diff

import (
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"strings"
)

func (c *Compare) CompareChapters() ([]Pair, string, *log.Status) {
	var fileMap string
	var status *log.Status
	var scripts []db.Script
	scripts, status = c.database.SelectBookChapter()
	if status != nil {
		return c.results, fileMap, status
	}
	for _, scp := range scripts {
		var pair Pair
		pair.Ref.BookId = scp.BookId
		pair.Ref.ChapterNum = scp.ChapterNum
		var baseText, text string
		baseText, status = c.concatText(c.baseDb, scp.BookId, scp.ChapterNum)
		if status != nil {
			return c.results, fileMap, status
		}
		pair.Base.Text = c.cleanup(baseText)
		pair.Base.ScriptId = scp.ScriptId
		text, status = c.concatText(c.database, scp.BookId, scp.ChapterNum)
		if status != nil {
			return c.results, fileMap, status
		}
		pair.Comp.Text = c.cleanup(text)
		pair.Comp.ScriptId = scp.ScriptId
		c.chapterDiff(pair)
	}
	fileMap, status = c.generateBookChapterFilenameMap()
	c.baseDb.Close()
	return c.results, fileMap, status
}

func (c *Compare) concatText(conn db.DBAdapter, bookId string, chapter int) (string, *log.Status) {
	var results []string
	scripts, status := conn.SelectScriptsByChapter(bookId, chapter)
	if status != nil {
		return ``, status
	}
	var priorText string
	for _, scp := range scripts {
		if !strings.HasSuffix(priorText, ` `) && !strings.HasPrefix(scp.ScriptText, ` `) {
			results = append(results, ` `)
		}
		results = append(results, scp.ScriptText)
		priorText = scp.ScriptText
	}
	return strings.Join(results, ""), status
}

func (c *Compare) chapterDiff(chapter Pair) {
	diffs := c.diffMatch.DiffMain(chapter.Comp.Uroman, chapter.Base.Uroman, false)
	diffs = c.diffMatch.DiffCleanupMerge(diffs)
	if !c.isMatch(diffs) {
		chapter.Diffs = diffs
		chapter.HTML = c.diffMatch.DiffPrettyHtml(diffs)
		c.results = append(c.results, chapter)
	}
}
