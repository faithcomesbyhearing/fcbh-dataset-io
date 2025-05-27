package encode

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/cli_misc"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/read"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// This unfinished code was an attempt to organize scripts by verse in order
// to use the script timestamps.  Then the remaining verses whose timestamp
// had not be set by the script, were to be given a timestamps from Aeneas.
// Sandeep thought this seemed like a good idea.  Jon thought the audio was too
// subject to revision for the script timestamps to be trusted.
//
//
// This class is not yet finished.  It cannot be deployed as part of this service
// because there is no consistent repository of the audio data.
// In order to get the aeneas to work, I would need to use start stop
// start_time = 10.0  # Start aligning from 10 seconds
// stop_time = 30.0  # Stop aligning at 30 seconds
// # Prepare the path for Aeneas to work with the files
// config_string =f"task_language=eng|os_task_file_code=path/to/output/path|is_text_file=true|start={start_time}|stop={stop_time}"
// config = aeneas.tools.prepare_path(config_string)

// 1. Load the script into a memory database
// 2. Parsing the script, find those script lines that begin a verse.
// 3. Break the script up into parts that contain 1 or more whole verse
// 4. Load the plain_text_edit script
// 5. Iterate over the script groups
// 6. Access the script timestamps each group
// 7. Set timestamps in grouped script
// 8. For each group get the corresponding verses from plain_text_edit
// 9. Process verses through aeneas
// 10. Compute the final timestamps of all verses
// 11. Update plain_text_edit timestamps

type AeneasExperiment struct {
	ctx          context.Context
	ts           cli_misc.TSBucket
	bibleId      string
	audioMediaId string
	languageISO  string
	textConn     db.DBAdapter
	//scriptIdMap  map[string]int
	audioFiles map[string]string
}

func NewAeneasExperiment(ctx context.Context, audioMediaId string, language string) AeneasExperiment {
	var a AeneasExperiment
	a.ctx = ctx
	a.ts = cli_misc.NewTSBucket(a.ctx)
	a.bibleId = audioMediaId[:6]
	a.audioMediaId = audioMediaId
	a.languageISO = language
	return a
}

func (a *AeneasExperiment) Process() {
	a.audioFiles = a.FindAudioFiles()
	script := a.LoadScript()
	fmt.Println(`script lines`, len(script))
	a.textConn = a.LoadPlainTextEdit()
	a.ProcessByChapter(script)
}

func (a *AeneasExperiment) FindAudioFiles() map[string]string {
	filePath := filepath.Join(os.Getenv(`FCBH_DATASET_FILES`), a.bibleId, a.audioMediaId, `*.mp3`)
	//testament := request.Testament{NT: true}
	files, status := input.FileInput(a.ctx, filePath)
	if status != nil {
		panic(status)
	}
	var result = make(map[string]string)
	for _, file := range files {
		key := file.BookId + strconv.Itoa(file.Chapter)
		result[key] = file.FilePath()
	}
	return result
}

func (a *AeneasExperiment) LoadScript() []db.Script {
	var results []db.Script
	var status *log.Status
	//ts := input.NewTSBucket(a.ctx)
	key := a.ts.GetKey(cli_misc.Script, a.audioMediaId, ``, 0)
	fmt.Println(`key:`, key)
	filePath := filepath.Join(os.Getenv(`FCBH_DATASET_FILES`), a.bibleId, a.audioMediaId[:8]+`ST.xlsx`)
	a.ts.DownloadObject(cli_misc.TSBucketName, key, filePath)
	fmt.Println(`path`, filePath)
	var conn = db.NewDBAdapter(a.ctx, ":memory:")
	testament := request.Testament{OT: true, NT: true}
	reader := read.NewScriptReader(conn, testament)
	reader.Read(filePath)
	results, status = conn.SelectScripts()
	if status != nil {
		panic(status)
	}
	//a.scriptIdMap, status = conn.SelectScriptIdScriptNum()
	//if status.IsErr {
	//	panic(status)
	//}
	conn.Close()
	fmt.Println(`count:`, len(results))
	return results
}

func (a *AeneasExperiment) LoadPlainTextEdit() db.DBAdapter {
	user := request.GetTestUser()
	conn, status := db.NewerDBAdapter(a.ctx, true, user, `Plain_Text_Edit_`+a.bibleId)
	if status != nil {
		panic(status)
	}
	var req request.Request
	req.BibleId = a.bibleId
	req.Testament = request.Testament{NT: true}
	reader := read.NewDBPTextEditReader(conn, req)
	reader.Process()
	return conn
}

func (a *AeneasExperiment) ProcessByChapter(scripts []db.Script) {
	var results []db.Script
	var lastBookId = ``
	var lastChapter = -1
	for _, scp := range scripts {
		if scp.BookId != lastBookId || scp.ChapterNum != lastChapter {
			if len(results) > 0 {
				results = a.GetTimestamps(a.audioMediaId, results)
				grouped := a.GroupScriptByVerse(results)
				a.GetPlainTextEdit(a.textConn, grouped)
			}
			results = nil
			lastBookId = scp.BookId
			lastChapter = scp.ChapterNum
		}
		results = append(results, scp)
	}
	if len(results) > 0 {
		results = a.GetTimestamps(a.audioMediaId, results)
		grouped := a.GroupScriptByVerse(results)
		a.GetPlainTextEdit(a.textConn, grouped)
	}
}

func (a *AeneasExperiment) GetTimestamps(audioId string, scripts []db.Script) []db.Script {
	timestamps := a.ts.GetTimestamps(cli_misc.ScriptTS, audioId, scripts[0].BookId, scripts[0].ChapterNum)
	//fmt.Println(`timestamps:`, timestamps)
	var timeMap = make(map[string]db.Timestamp)
	for _, t := range timestamps {
		timeMap[t.VerseStr] = t
	}
	for i, scp := range scripts {
		scripts[i].ScriptBeginTS = -1
		scripts[i].ScriptEndTS = -1
		ts, ok := timeMap[scp.ScriptNum]
		if ok {
			scripts[i].ScriptBeginTS = ts.BeginTS
			scripts[i].ScriptEndTS = ts.EndTS
		}
	}
	// Repair missing timestamps where possible
	for i, _ := range scripts {
		if scripts[i].ScriptBeginTS == -1 && i > 0 {
			scripts[i].ScriptBeginTS = scripts[i-1].ScriptEndTS
		}
		if scripts[i].ScriptEndTS == -1 && i < len(scripts)-1 {
			scripts[i].ScriptEndTS = scripts[i+1].ScriptBeginTS
		}
	}
	for _, scp := range scripts {
		if scp.ScriptBeginTS == -1 || scp.ScriptEndTS == -1 {
			ref := scp.BookId + ` ` + strconv.Itoa(scp.ChapterNum) + `:` + scp.VerseStr + ` (` + scp.ScriptNum + `)`
			log.Warn(a.ctx, `No timestamp found for verse `, ref)
		}
	}
	return scripts
}

func (a *AeneasExperiment) GroupScriptByVerse(scripts []db.Script) []db.Script {
	var results = make([]db.Script, 0, len(scripts))
	var rec db.Script
	for _, scp := range scripts {
		trimmed := strings.TrimSpace(scp.ScriptText)
		if trimmed[0] == '{' || scp.VerseStr == `0` {
			if rec.BookId != `` {
				results = append(results, rec)
			}
			rec = scp
		}
		rec.VerseEnd = scp.VerseStr
		rec.ScriptEndTS = scp.ScriptEndTS
		rec.ScriptTexts = append(rec.ScriptTexts, scp.ScriptText)
	}
	if rec.BookId != `` {
		results = append(results, rec)
	}
	for i, scp := range results {
		results[i].ScriptText = strings.Join(scp.ScriptTexts, "")
	}
	return results
}

func (a *AeneasExperiment) GetPlainTextEdit(conn db.DBAdapter, groups []db.Script) {
	// scripts here is missing script_id do I need it?
	scripts, status := conn.SelectScriptsByChapter(groups[0].BookId, groups[0].ChapterNum)
	if status != nil {
		panic(status)
	}
	var beginMap = make(map[string]float64)
	var endMap = make(map[string]float64)
	for _, grp := range groups {
		beginMap[grp.VerseStr] = grp.ScriptBeginTS
		endMap[grp.VerseEnd] = grp.ScriptEndTS
	}
	for i, scp := range scripts {
		scripts[i].ScriptBeginTS = -1.0
		scripts[i].ScriptEndTS = -1.0
		ts, ok := beginMap[scp.VerseStr]
		if ok {
			scripts[i].ScriptBeginTS = ts
		}
		ts, ok = endMap[scp.VerseStr]
		if ok {
			scripts[i].ScriptEndTS = ts
		}
	}
	var allAeneas []db.Timestamp
	for _, part := range groups {
		if part.VerseStr != part.VerseEnd {
			verseStr := a.SafeAtoi(part.VerseStr)
			verseEnd := a.SafeAtoi(part.VerseEnd)
			var recs []db.Script
			for _, scp := range scripts {
				verseNum := a.SafeAtoi(scp.VerseStr)
				if verseNum >= verseStr && verseNum <= verseEnd {
					recs = append(recs, scp)
				}
			}
			if len(recs) > 0 {
				aeneasTS := a.ProcessAeneas(recs)
				allAeneas = append(allAeneas, aeneasTS...)
				recs = nil
			}
		}
	}
	a.MergeTimestamps(scripts, allAeneas)
}

func (a *AeneasExperiment) SafeAtoi(number string) int {
	var result []rune
	for _, chr := range number {
		if chr >= '0' && chr <= '9' {
			result = append(result, chr)
		} else {
			break
		}
	}
	num, _ := strconv.Atoi(string(result))
	return num
}

func (a *AeneasExperiment) ProcessAeneas(recs []db.Script) []db.Timestamp {
	ref := recs[0].BookId + ` ` + strconv.Itoa(recs[0].ChapterNum) + `_*.txt`
	var fp, err = os.CreateTemp(os.Getenv(`FCBH_DATASET_TMP`), ref)
	if err != nil {
		panic(err)
	}
	for _, rec := range recs {
		_, _ = fp.WriteString(rec.VerseStr)
		_, _ = fp.WriteString("|")
		_, _ = fp.WriteString(rec.ScriptText)
		if !strings.HasSuffix(rec.ScriptText, "\n") {
			_, _ = fp.WriteString("\n")
		}
	}
	_ = fp.Close()
	content, _ := os.ReadFile(fp.Name())
	fmt.Println(string(content))
	aeneas := NewAeneas(a.ctx, a.textConn, a.bibleId, `epo`, request.Detail{Lines: true})
	key := recs[0].BookId + strconv.Itoa(recs[0].ChapterNum)
	audioFile, ok := a.audioFiles[key]
	if !ok {
		panic("audio file not found")
	}
	filename, status := aeneas.executeAeneas(a.languageISO, audioFile, fp.Name())
	if status != nil {
		panic(status)
	}
	timestamps, status := aeneas.parseResponse(filename, audioFile)
	if status != nil {
		panic(status)
	}
	for i, _ := range timestamps {
		timestamps[i].VerseStr = strconv.Itoa(timestamps[i].Id)
	}
	return timestamps
}

func (a *AeneasExperiment) MergeTimestamps(scripts []db.Script, aeneasTS []db.Timestamp) {
	var results []db.Script
	var scrIdx = 0
	var aenIdx = 0
	var beginTS = 0.0
	for {
		if scrIdx >= len(scripts) && aenIdx >= len(aeneasTS) {
			break
		}
		var script db.Script
		var aeneas db.Timestamp
		var scrVerse = 10000
		var aenVerse = 10000
		if scrIdx < len(scripts) {
			script = scripts[scrIdx]
			scrVerse = a.SafeAtoi(script.VerseStr)
		}
		if aenIdx < len(aeneasTS) {
			aeneas = aeneasTS[aenIdx]
			aenVerse = a.SafeAtoi(aeneas.VerseStr)
		}
		if scrVerse < aenVerse {
			results = append(results, script)
			beginTS = script.ScriptEndTS
			scrIdx++
		} else if scrVerse > aenVerse {
			var aen db.Script
			aen.VerseStr = aeneas.VerseStr
			aen.ScriptBeginTS = aeneas.BeginTS + beginTS
			aen.ScriptEndTS = aeneas.EndTS + beginTS
			results = append(results, aen)
			beginTS = aen.ScriptEndTS
			aenIdx++
		} else { // Are equal
			if script.ScriptBeginTS == -1 {
				script.ScriptBeginTS = aeneas.BeginTS + beginTS
			}
			if script.ScriptEndTS == -1 {
				script.ScriptEndTS = aeneas.EndTS + beginTS
			}
			results = append(results, script)
			beginTS = script.ScriptEndTS
			scrIdx++
			aenIdx++
		}
	}
	for _, ts := range results {
		fmt.Println(ts)
	}
}

func (a *AeneasExperiment) CountChars(scripts []db.Script) int {
	var results int
	for _, scp := range scripts {
		results += len(scp.ScriptText)
	}
	return results
}

func (a *AeneasExperiment) DumpScript(scripts []db.Script, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	for _, scp := range scripts {
		//_, err = fmt.Fprintln(file, scp.BookId, scp.ChapterNum, scp.VerseStr, scp.ScriptNum, scp.ScriptText)
		_, err = fmt.Fprintln(file, scp.ScriptText)
		if err != nil {
			panic(err)
		}
	}
}
