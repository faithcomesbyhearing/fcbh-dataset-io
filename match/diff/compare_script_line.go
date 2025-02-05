package diff

import (
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

func (c *Compare) CompareScriptLines() ([]Pair, string, *log.Status) {
	var fileMap string
	var status *log.Status
	var scripts []db.Script
	scripts, status = c.database.SelectScripts()
	if status != nil {
		return c.results, fileMap, status
	}
	var compareMap = make(map[string]db.Script)
	for _, comp := range scripts {
		compareMap[comp.ScriptNum] = comp
	}
	// Select and increment over the base scripts
	scripts, status = c.baseDb.SelectScripts()
	if status != nil {
		return c.results, fileMap, status
	}
	for _, base := range scripts {
		baseText := c.cleanup(base.ScriptText)
		if c.baseIdent.TextSource == request.TextScript {
			baseText = c.verseRm.ReplaceAllString(baseText, ``)
		}
		var pair Pair
		pair.Ref.BookId = base.BookId
		pair.Ref.ChapterNum = base.ChapterNum
		pair.Ref.VerseStr = base.ScriptNum
		comp, ok := compareMap[base.ScriptNum]
		if ok {
			compText := c.cleanup(comp.ScriptText)
			if c.compIdent.TextSource == request.TextScript {
				compText = c.verseRm.ReplaceAllString(compText, ``)
			}
			pair.Base.Text = baseText
			pair.Comp.Text = compText
			c.scriptLineDiff(pair)
			delete(compareMap, base.ScriptNum)
		} else {
			pair.Base.Text = baseText
			pair.Comp.Text = ""
			c.scriptLineDiff(pair)
		}
	}
	// Report any compare scripts that had no entry in base
	for _, comp := range compareMap {
		compText := c.cleanup(comp.ScriptText)
		if c.compIdent.TextSource == request.TextScript {
			compText = c.verseRm.ReplaceAllString(compText, ``)
		}
		var pair Pair
		pair.Ref.BookId = comp.BookId
		pair.Ref.ChapterNum = comp.ChapterNum
		pair.Ref.VerseStr = comp.ScriptNum
		pair.Base.Text = ""
		pair.Comp.Text = compText
		c.scriptLineDiff(pair)
	}
	fileMap, status = c.generateBookChapterFilenameMap()
	c.baseDb.Close()
	return c.results, fileMap, status
}

func (c *Compare) scriptLineDiff(script Pair) {
	diffs := c.diffMatch.DiffMain(script.Comp.Text, script.Base.Text, false)
	diffs = c.diffMatch.DiffCleanupMerge(diffs)
	if !c.isMatch(diffs) {
		script.Diffs = diffs
		script.HTML = c.diffMatch.DiffPrettyHtml(diffs)
		c.results = append(c.results, script)
	}
}
