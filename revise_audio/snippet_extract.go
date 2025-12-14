package revise_audio

import (
	"context"
	"os"
	"path/filepath"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/ffmpeg"
)

// SnippetExtractor handles extraction of audio snippets from chapter files
type SnippetExtractor struct {
	ctx     context.Context
	conn    db.DBAdapter
	tempDir string
	basePath string // Base path for audio files (from FCBH_DATASET_FILES)
}

// NewSnippetExtractor creates a new snippet extractor
func NewSnippetExtractor(ctx context.Context, conn db.DBAdapter, basePath string) *SnippetExtractor {
	return &SnippetExtractor{
		ctx:      ctx,
		conn:     conn,
		basePath: basePath,
	}
}

// ExtractSnippet extracts an audio snippet based on a SnippetCandidate
// Returns the path to the extracted audio file (WAV format)
func (s *SnippetExtractor) ExtractSnippet(candidate SnippetCandidate) (string, *log.Status) {
	if len(candidate.Words) == 0 {
		return "", log.ErrorNoErr(s.ctx, 400, "No words in candidate snippet")
	}

	// Resolve audio file path
	audioFile, status := s.resolveAudioFilePath(candidate)
	if status != nil {
		return "", status
	}

	// Create temp directory if needed
	if s.tempDir == "" {
		var err error
		s.tempDir, err = os.MkdirTemp(os.Getenv("FCBH_DATASET_TMP"), "revise_audio_")
		if err != nil {
			return "", log.Error(s.ctx, 500, err, "Error creating temp directory")
		}
	}

	// Extract snippet using timestamps from first and last word
	startTS := candidate.StartTS
	endTS := candidate.EndTS

	// Use ffmpeg to extract the segment
	outputFile, status := ffmpeg.ChopOneSegment(s.ctx, s.tempDir, audioFile, startTS, endTS)
	if status != nil {
		return "", status
	}

	return outputFile, nil
}

// ExtractWord extracts a single word from the database
// This is a convenience method that queries the database and extracts the word
// marginSeconds: optional additional time before and after the word (default 0.0)
func (s *SnippetExtractor) ExtractWord(bookId string, chapterNum int, verseStr string, wordSeq int, marginSeconds ...float64) (string, *log.Status) {
	margin := 0.0 // Default no margin
	if len(marginSeconds) > 0 && marginSeconds[0] > 0 {
		margin = marginSeconds[0]
	}

	// Query database for words with timestamps
	words, status := s.queryWordsByVerse(bookId, chapterNum, verseStr)
	if status != nil {
		return "", status
	}

	// Find the specific word
	var targetWord *db.Audio
	for i := range words {
		if words[i].WordSeq == wordSeq {
			targetWord = &words[i]
			break
		}
	}

	if targetWord == nil {
		return "", log.ErrorNoErr(s.ctx, 404, "Word not found", bookId, chapterNum, verseStr, wordSeq)
	}

	// Resolve audio file path (already in targetWord from the query)
	audioFile, status := s.resolveAudioFilePathString(targetWord.AudioFile)
	if status != nil {
		return "", status
	}

	// Create temp directory if needed
	if s.tempDir == "" {
		var err error
		s.tempDir, err = os.MkdirTemp(os.Getenv("FCBH_DATASET_TMP"), "revise_audio_")
		if err != nil {
			return "", log.Error(s.ctx, 500, err, "Error creating temp directory")
		}
	}

	// Extract snippet with margin
	beginTS := targetWord.BeginTS - margin
	if beginTS < 0 {
		beginTS = 0
	}
	endTS := targetWord.EndTS + margin

	outputFile, status := ffmpeg.ChopOneSegment(s.ctx, s.tempDir, audioFile, beginTS, endTS)
	if status != nil {
		return "", status
	}

	return outputFile, nil
}

// ExtractPhrase extracts a phrase (multiple consecutive words) from the database
func (s *SnippetExtractor) ExtractPhrase(bookId string, chapterNum int, verseStr string, startWordSeq int, endWordSeq int) (string, *log.Status) {
	// Query database for words with timestamps
	words, status := s.queryWordsByVerse(bookId, chapterNum, verseStr)
	if status != nil {
		return "", status
	}

	// Find words in range
	var phraseWords []db.Audio
	for i := range words {
		if words[i].WordSeq >= startWordSeq && words[i].WordSeq <= endWordSeq {
			phraseWords = append(phraseWords, words[i])
		}
	}

	if len(phraseWords) == 0 {
		return "", log.ErrorNoErr(s.ctx, 404, "No words found in range", bookId, chapterNum, verseStr, startWordSeq, endWordSeq)
	}

	// Resolve audio file path (already in phraseWords from the query)
	audioFile, status := s.resolveAudioFilePathString(phraseWords[0].AudioFile)
	if status != nil {
		return "", status
	}

	// Create temp directory if needed
	if s.tempDir == "" {
		var err error
		s.tempDir, err = os.MkdirTemp(os.Getenv("FCBH_DATASET_TMP"), "revise_audio_")
		if err != nil {
			return "", log.Error(s.ctx, 500, err, "Error creating temp directory")
		}
	}

	// Extract snippet from first word start to last word end
	startTS := phraseWords[0].BeginTS
	endTS := phraseWords[len(phraseWords)-1].EndTS

	outputFile, status := ffmpeg.ChopOneSegment(s.ctx, s.tempDir, audioFile, startTS, endTS)
	if status != nil {
		return "", status
	}

	return outputFile, nil
}

// Cleanup removes temporary files
func (s *SnippetExtractor) Cleanup() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
		s.tempDir = ""
	}
}

// resolveAudioFilePath resolves the full path to an audio file from a candidate
func (s *SnippetExtractor) resolveAudioFilePath(candidate SnippetCandidate) (string, *log.Status) {
	// Get script to find audio file path
	scripts, status := s.conn.SelectScriptsByBookChapter(candidate.BookId, candidate.ChapterNum)
	if status != nil {
		return "", status
	}
	
	var script *db.Script
	for i := range scripts {
		if scripts[i].ScriptId == candidate.ScriptId {
			script = &scripts[i]
			break
		}
	}
	
	if script == nil {
		return "", log.ErrorNoErr(s.ctx, 404, "Script not found", candidate.ScriptId)
	}

	return s.resolveAudioFilePathFromScript(*script)
}

// queryWordsByVerse queries words with timestamps for a specific verse
func (s *SnippetExtractor) queryWordsByVerse(bookId string, chapterNum int, verseStr string) ([]db.Audio, *log.Status) {
	query := `SELECT w.word_id, w.script_id, s.book_id, s.chapter_num, s.verse_str, 
		s.verse_num, w.word_seq, w.word, w.uroman, w.word_begin_ts, w.word_end_ts, w.fa_score,
		s.script_begin_ts, s.script_end_ts, s.fa_score, s.audio_file
		FROM words w JOIN scripts s ON w.script_id = s.script_id
		WHERE w.ttype = 'W' AND s.book_id = ? AND s.chapter_num = ? AND s.verse_str = ?
		ORDER BY w.word_seq`
	
	rows, err := s.conn.DB.Query(query, bookId, chapterNum, verseStr)
	if err != nil {
		return nil, log.Error(s.ctx, 500, err, "Error querying words by verse")
	}
	defer rows.Close()
	
	var results []db.Audio
	for rows.Next() {
		var rec db.Audio
		err := rows.Scan(
			&rec.WordId, &rec.ScriptId, &rec.BookId, &rec.ChapterNum, &rec.VerseStr,
			&rec.VerseSeq, &rec.WordSeq, &rec.Text, &rec.Uroman,
			&rec.BeginTS, &rec.EndTS, &rec.FAScore,
			&rec.ScriptBeginTS, &rec.ScriptEndTS, &rec.ScriptFAScore, &rec.AudioFile,
		)
		if err != nil {
			return nil, log.Error(s.ctx, 500, err, "Error scanning word row")
		}
		results = append(results, rec)
	}
	
	if err := rows.Err(); err != nil {
		return nil, log.Error(s.ctx, 500, err, "Error iterating word rows")
	}
	
	return results, nil
}

// resolveAudioFilePathFromScript resolves the full path from a script record
func (s *SnippetExtractor) resolveAudioFilePathFromScript(script db.Script) (string, *log.Status) {
	return s.resolveAudioFilePathString(script.AudioFile)
}

// resolveAudioFilePathString resolves the full path from an audio file string
// The database may store just the filename, so we search in subdirectories if needed
func (s *SnippetExtractor) resolveAudioFilePathString(audioFile string) (string, *log.Status) {
	if audioFile == "" {
		return "", log.ErrorNoErr(s.ctx, 400, "Audio file path is empty")
	}

	// If path is already absolute, use it
	if filepath.IsAbs(audioFile) {
		if _, err := os.Stat(audioFile); err == nil {
			return audioFile, nil
		}
		// If absolute path doesn't exist, continue to search
	}

	// Get base path
	if s.basePath == "" {
		s.basePath = os.Getenv("FCBH_DATASET_FILES")
		if s.basePath == "" {
			return "", log.ErrorNoErr(s.ctx, 400, "FCBH_DATASET_FILES not set and basePath not provided")
		}
	}

	// Try direct path first
	fullPath := filepath.Join(s.basePath, audioFile)
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath, nil
	}

	// If not found, search in subdirectories (database may store just filename)
	filename := filepath.Base(audioFile)
	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue searching
		}
		if !info.IsDir() && info.Name() == filename {
			fullPath = path
			return filepath.SkipAll // Found it, stop searching
		}
		return nil
	})
	
	if err == nil && fullPath != filepath.Join(s.basePath, audioFile) {
		// Found in subdirectory
		return fullPath, nil
	}

	// Not found anywhere
	return "", log.ErrorNoErr(s.ctx, 404, "Audio file not found", audioFile, "searched in", s.basePath)
}

