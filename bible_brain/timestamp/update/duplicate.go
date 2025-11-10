package update

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

type durationMismatch struct {
	BookID         string
	Chapter        int
	SourceDuration float64
	TargetDuration float64
	Reason         string
}

type durationComparison struct {
	Chapters   []db.Script
	Mismatches []durationMismatch
}

func (d *UpdateTimestamps) handleDuplicationIfNeeded(ident db.Ident) (bool, []db.Script, *log.Status) {
	targetID := strings.TrimSpace(d.req.UpdateDBP.Timestamps)
	if targetID == "" {
		return false, nil, nil
	}

	sourceID := strings.TrimSpace(d.req.UpdateDBP.CopyTimestampsFrom)
	if sourceID == "" {
		sourceID = inferSourceFileset(ident, targetID)
	}
	if sourceID == "" || strings.EqualFold(sourceID, targetID) {
		return false, nil, nil
	}

	sourceID = strings.ToUpper(sourceID)
	targetID = strings.ToUpper(targetID)

	tolerance := duplicationTolerance()

	sourceDurations, status := d.dbpConn.GetFilesetDurations(sourceID)
	if status != nil {
		return false, nil, status
	}
	if len(sourceDurations) == 0 {
		return false, nil, log.ErrorNoErr(d.ctx, 400, fmt.Sprintf("Source fileset %s has no duration tags available", sourceID))
	}

	targetDurations, status := d.dbpConn.GetFilesetDurations(targetID)
	if status != nil {
		return false, nil, status
	}
	if len(targetDurations) == 0 {
		return false, nil, log.ErrorNoErr(d.ctx, 400, fmt.Sprintf("Target fileset %s has no duration tags available", targetID))
	}

	comparison := compareDurations(sourceDurations, targetDurations, tolerance)
	if len(comparison.Chapters) == 0 {
		return false, nil, log.ErrorNoErr(d.ctx, 400, fmt.Sprintf("No matching chapters between %s and %s", sourceID, targetID))
	}
	if len(comparison.Mismatches) > 0 {
		return false, nil, log.ErrorNoErr(d.ctx, 422, formatMismatchError(sourceID, targetID, comparison.Mismatches))
	}

	timestampData, chapters, status := d.dbpConn.GetFilesetTimestamps(sourceID)
	if status != nil {
		return false, nil, status
	}

	filteredData, filteredChapters := filterTimestampData(timestampData, chapters, comparison.Chapters)
	if len(filteredChapters) == 0 {
		return false, nil, log.ErrorNoErr(d.ctx, 400, fmt.Sprintf("No timestamp data available to duplicate from %s to %s", sourceID, targetID))
	}

	resetTimestampIDs(filteredData)

	status = d.dbpConn.ProcessTimestamps(targetID, mmsAlignTimingEstErr, filteredChapters, filteredData)
	if status != nil {
		return false, nil, status
	}

	log.Info(d.ctx, "Duplicated timestamps from", sourceID, "to", targetID, "chapters:", len(filteredChapters))
	return true, filteredChapters, nil
}

func compareDurations(source, target map[string]map[int]float64, tolerance float64) durationComparison {
	allowed := make(map[string]map[int]bool)
	mismatches := make([]durationMismatch, 0)
	chapters := make([]db.Script, 0)

	for bookID, chapterMap := range source {
		for chapter, srcDuration := range chapterMap {
			tgtDuration, ok := lookupDuration(target, bookID, chapter)
			if !ok {
				mismatches = append(mismatches, durationMismatch{
					BookID:         bookID,
					Chapter:        chapter,
					SourceDuration: srcDuration,
					TargetDuration: 0,
					Reason:         "missing target duration",
				})
				continue
			}

			if math.Abs(srcDuration-tgtDuration) > tolerance {
				mismatches = append(mismatches, durationMismatch{
					BookID:         bookID,
					Chapter:        chapter,
					SourceDuration: srcDuration,
					TargetDuration: tgtDuration,
					Reason:         "duration mismatch",
				})
				continue
			}

			if _, ok := allowed[bookID]; !ok {
				allowed[bookID] = make(map[int]bool)
			}
			allowed[bookID][chapter] = true
			chapters = append(chapters, db.Script{BookId: bookID, ChapterNum: chapter})
		}
	}

	return durationComparison{
		Chapters:   chapters,
		Mismatches: mismatches,
	}
}

func resetTimestampIDs(data map[string]map[int][]Timestamp) {
	for _, chapterMap := range data {
		for chapter := range chapterMap {
			for i := range chapterMap[chapter] {
				// insertTimestampsTx only inserts rows whose TimestampId == 0.
				// Zero it out here so copied rows are treated as new inserts.
				chapterMap[chapter][i].TimestampId = 0
			}
		}
	}
}

func filterTimestampData(data map[string]map[int][]Timestamp, existing []db.Script, allowed []db.Script) (map[string]map[int][]Timestamp, []db.Script) {
	if len(allowed) == 0 {
		return map[string]map[int][]Timestamp{}, nil
	}

	allowedMap := make(map[string]map[int]bool)
	for _, ch := range allowed {
		if _, ok := allowedMap[ch.BookId]; !ok {
			allowedMap[ch.BookId] = make(map[int]bool)
		}
		allowedMap[ch.BookId][ch.ChapterNum] = true
	}

	result := make(map[string]map[int][]Timestamp)
	filteredChapters := make([]db.Script, 0, len(existing))

	for _, ch := range existing {
		if allowedMap[ch.BookId][ch.ChapterNum] {
			if _, ok := data[ch.BookId]; !ok {
				continue
			}
			if _, ok := data[ch.BookId][ch.ChapterNum]; !ok {
				continue
			}
			if _, ok := result[ch.BookId]; !ok {
				result[ch.BookId] = make(map[int][]Timestamp)
			}
			result[ch.BookId][ch.ChapterNum] = data[ch.BookId][ch.ChapterNum]
			filteredChapters = append(filteredChapters, ch)
		}
	}

	return result, filteredChapters
}

func lookupDuration(source map[string]map[int]float64, bookID string, chapter int) (float64, bool) {
	bookMap, ok := source[bookID]
	if !ok {
		return 0, false
	}
	duration, ok := bookMap[chapter]
	return duration, ok
}

func formatMismatchError(sourceID, targetID string, mismatches []durationMismatch) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Duration mismatch between %s and %s", sourceID, targetID))

	maxExamples := 5
	for i, mismatch := range mismatches {
		if i == 0 {
			builder.WriteString(": ")
		}
		if i >= maxExamples {
			builder.WriteString(fmt.Sprintf("... (%d more)", len(mismatches)-maxExamples))
			break
		}
		builder.WriteString(fmt.Sprintf("%s %d src=%.2fs tgt=%.2fs (%s)", mismatch.BookID, mismatch.Chapter, mismatch.SourceDuration, mismatch.TargetDuration, mismatch.Reason))
		if i < len(mismatches)-1 && i < maxExamples-1 {
			builder.WriteString("; ")
		}
	}

	return builder.String()
}

func duplicationTolerance() float64 {
	value := strings.TrimSpace(os.Getenv("BB_DUPLICATION_TOLERANCE"))
	if value == "" {
		return 0
	}

	tolerance, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	if tolerance < 0 {
		return 0
	}
	return tolerance
}

func inferSourceFileset(ident db.Ident, targetID string) string {
	upper := strings.ToUpper(targetID)
	switch {
	case strings.Contains(upper, "N2"):
		if ident.AudioNTId != "" {
			return ident.AudioNTId
		}
	case strings.Contains(upper, "O2"):
		if ident.AudioOTId != "" {
			return ident.AudioOTId
		}
	}
	return ""
}
