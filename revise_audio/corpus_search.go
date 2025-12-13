package revise_audio

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

// CorpusSearcher handles searching the corpus for replacement audio snippets
type CorpusSearcher struct {
	ctx      context.Context
	conn     db.DBAdapter
	basePath string // FCBH_DATASET_FILES base path
}

// NewCorpusSearcher creates a new CorpusSearcher instance
func NewCorpusSearcher(ctx context.Context, conn db.DBAdapter) *CorpusSearcher {
	basePath := os.Getenv("FCBH_DATASET_FILES")
	if basePath == "" {
		basePath = os.Getenv("HOME") + "/tmp/arti/files" // fallback
	}
	return &CorpusSearcher{
		ctx:      ctx,
		conn:     conn,
		basePath: basePath,
	}
}

// FindReplacementSnippets searches the corpus for audio snippets matching the replacement text
// Returns the best match by default, or all candidates if returnAll is true
func (c *CorpusSearcher) FindReplacementSnippets(
	replacementText string,
	targetBookId string,
	targetChapterNum int,
	targetVerseNum int,
	returnAll bool,
) ([]SnippetCandidate, *log.Status) {
	// Normalize the replacement text
	normalizedText := normalizeText(replacementText)
	if normalizedText == "" {
		return nil, log.ErrorNoErr(c.ctx, 400, "Replacement text is empty")
	}

	// Split into words for phrase matching
	words := strings.Fields(normalizedText)
	if len(words) == 0 {
		return nil, log.ErrorNoErr(c.ctx, 400, "Replacement text has no words")
	}

	// Find all candidates
	candidates, status := c.findCandidates(normalizedText, words, targetBookId, targetChapterNum, targetVerseNum)
	if status != nil {
		return nil, status
	}

	if len(candidates) == 0 {
		// Return empty list (not an error - user can review manually)
		return []SnippetCandidate{}, nil
	}

	// Rank candidates
	ranked := rankCandidates(candidates, targetBookId, targetChapterNum, targetVerseNum)

	if returnAll {
		return ranked, nil
	}

	// Return best match
	if len(ranked) > 0 {
		return []SnippetCandidate{ranked[0]}, nil
	}

	return []SnippetCandidate{}, nil
}

// findCandidates searches the database for matching words/phrases
func (c *CorpusSearcher) findCandidates(
	normalizedText string,
	words []string,
	targetBookId string,
	targetChapterNum int,
	targetVerseNum int,
) ([]SnippetCandidate, *log.Status) {
	var candidates []SnippetCandidate

	// Strategy: First try to find whole phrase matches, then fall back to word-by-word
	// Pass 1: Find whole phrase matches (exact text match)
	phraseCandidates, status := c.findPhraseMatches(normalizedText, targetBookId, targetChapterNum, targetVerseNum)
	if status != nil {
		return nil, status
	}
	candidates = append(candidates, phraseCandidates...)

	// If we found whole phrase matches, prefer those (will be ranked higher)
	// But also collect word-by-word matches for fallback
	if len(phraseCandidates) == 0 {
		// Pass 2: Find word-by-word matches and assemble phrases
		wordCandidates, status := c.findWordMatches(words, targetBookId, targetChapterNum, targetVerseNum)
		if status != nil {
			return nil, status
		}
		candidates = append(candidates, wordCandidates...)
	}

	return candidates, nil
}

// findPhraseMatches searches for exact phrase matches in the corpus
func (c *CorpusSearcher) findPhraseMatches(
	normalizedText string,
	targetBookId string,
	targetChapterNum int,
	targetVerseNum int,
) ([]SnippetCandidate, *log.Status) {
	var candidates []SnippetCandidate

	// Query for words matching the first word of the phrase
	// Then check if consecutive words form the full phrase
	query := `SELECT w.word_id, w.script_id, s.book_id, s.chapter_num, s.verse_str, 
		s.verse_num, w.word_seq, w.word, w.uroman, w.word_begin_ts, w.word_end_ts, w.fa_score,
		s.script_begin_ts, s.script_end_ts, s.fa_score, s.audio_file, s.actor, s.person
		FROM words w JOIN scripts s ON w.script_id = s.script_id
		WHERE w.ttype = 'W' AND LOWER(TRIM(w.word)) = LOWER(?)
		ORDER BY s.book_id, s.chapter_num, s.verse_num, w.word_seq`

	phraseWords := strings.Fields(normalizedText)
	phraseLength := len(phraseWords)
	if phraseLength == 0 {
		return candidates, nil
	}

	firstWord := phraseWords[0]
	rows, err := c.conn.DB.Query(query, firstWord)
	if err != nil {
		return nil, log.Error(c.ctx, 500, err, "Error querying for phrase matches")
	}
	defer rows.Close()

	// Group words by script_id to check for consecutive sequences
	// Also store actor/person info per script
	type wordWithActor struct {
		audio  db.Audio
		actor  string
		person string
	}
	scriptWords := make(map[int64][]wordWithActor)
	scriptActors := make(map[int64]struct {
		actor  string
		person string
	})

	for rows.Next() {
		var rec db.Audio
		var actor, person string
		err := rows.Scan(
			&rec.WordId, &rec.ScriptId, &rec.BookId, &rec.ChapterNum, &rec.VerseStr,
			&rec.VerseSeq, &rec.WordSeq, &rec.Text, &rec.Uroman,
			&rec.BeginTS, &rec.EndTS, &rec.FAScore,
			&rec.ScriptBeginTS, &rec.ScriptEndTS, &rec.ScriptFAScore, &rec.AudioFile,
			&actor, &person,
		)
		if err != nil {
			log.Warn(c.ctx, err, "Error scanning phrase match row")
			continue
		}
		scriptWords[rec.ScriptId] = append(scriptWords[rec.ScriptId], wordWithActor{
			audio:  rec,
			actor:  actor,
			person: person,
		})
		scriptActors[rec.ScriptId] = struct {
			actor  string
			person string
		}{actor: actor, person: person}
	}

	if err := rows.Err(); err != nil {
		return nil, log.Error(c.ctx, 500, err, "Error iterating phrase match rows")
	}

	// For each script, check if we have a consecutive sequence matching the phrase
	for scriptId, wordList := range scriptWords {
		// Sort words by word_seq
		for i := 0; i < len(wordList); i++ {
			for j := i + 1; j < len(wordList); j++ {
				if wordList[i].audio.WordSeq > wordList[j].audio.WordSeq {
					wordList[i], wordList[j] = wordList[j], wordList[i]
				}
			}
		}

		// Check for consecutive sequences of the right length
		if len(wordList) < phraseLength {
			continue
		}

		for i := 0; i <= len(wordList)-phraseLength; i++ {
			sequence := wordList[i : i+phraseLength]
			// Check if words are consecutive (word_seq increments by 1)
			isConsecutive := true
			for j := 1; j < len(sequence); j++ {
				if sequence[j].audio.WordSeq != sequence[j-1].audio.WordSeq+1 {
					isConsecutive = false
					break
				}
			}
			if !isConsecutive {
				continue
			}

			// Extract Audio records for sequence
			audioSeq := make([]db.Audio, len(sequence))
			for j, w := range sequence {
				audioSeq[j] = w.audio
			}

			// Check if this sequence matches the phrase
			sequenceText := buildPhraseText(audioSeq)
			if normalizeText(sequenceText) == normalizedText {
				// Found a matching phrase!
				actorInfo := scriptActors[scriptId]
				candidate := c.buildSnippetCandidateWithActor(audioSeq, actorInfo.actor, actorInfo.person, targetBookId, targetChapterNum, targetVerseNum)
				candidates = append(candidates, candidate)
				break // Only take first match per script
			}
		}
	}

	return candidates, nil
}

// findWordMatches finds individual word matches and attempts to assemble phrases
// This is a fallback when whole phrase matches aren't found
func (c *CorpusSearcher) findWordMatches(
	words []string,
	targetBookId string,
	targetChapterNum int,
	targetVerseNum int,
) ([]SnippetCandidate, *log.Status) {
	var candidates []SnippetCandidate

	// For now, find the first word match as a simple candidate
	// TODO: Implement more sophisticated phrase assembly from multiple sources
	if len(words) == 0 {
		return candidates, nil
	}

	query := `SELECT w.word_id, w.script_id, s.book_id, s.chapter_num, s.verse_str, 
		s.verse_num, w.word_seq, w.word, w.uroman, w.word_begin_ts, w.word_end_ts, w.fa_score,
		s.script_begin_ts, s.script_end_ts, s.fa_score, s.audio_file, s.actor, s.person
		FROM words w JOIN scripts s ON w.script_id = s.script_id
		WHERE w.ttype = 'W' AND LOWER(TRIM(w.word)) = LOWER(?)
		ORDER BY s.book_id, s.chapter_num, s.verse_num, w.word_seq
		LIMIT 100`

	firstWord := words[0]
	rows, err := c.conn.DB.Query(query, firstWord)
	if err != nil {
		return nil, log.Error(c.ctx, 500, err, "Error querying for word matches")
	}
	defer rows.Close()

	for rows.Next() {
		var rec db.Audio
		var actor, person string
		err := rows.Scan(
			&rec.WordId, &rec.ScriptId, &rec.BookId, &rec.ChapterNum, &rec.VerseStr,
			&rec.VerseSeq, &rec.WordSeq, &rec.Text, &rec.Uroman,
			&rec.BeginTS, &rec.EndTS, &rec.FAScore,
			&rec.ScriptBeginTS, &rec.ScriptEndTS, &rec.ScriptFAScore, &rec.AudioFile,
			&actor, &person,
		)
		if err != nil {
			log.Warn(c.ctx, err, "Error scanning word match row")
			continue
		}

		// Convert Audio to Word for candidate
		wordRec := db.Word{
			WordId:      int(rec.WordId),
			ScriptId:    int(rec.ScriptId),
			WordSeq:     rec.WordSeq,
			VerseNum:    rec.VerseSeq,
			Word:        rec.Text,
			WordBeginTS: rec.BeginTS,
			WordEndTS:   rec.EndTS,
			FAScore:     rec.FAScore,
		}

		candidate := c.buildSnippetCandidateFromWord(wordRec, rec, actor, person, targetBookId, targetChapterNum, targetVerseNum)
		candidates = append(candidates, candidate)
	}

	if err := rows.Err(); err != nil {
		return nil, log.Error(c.ctx, 500, err, "Error iterating word match rows")
	}

	return candidates, nil
}

// buildSnippetCandidateWithActor creates a SnippetCandidate from a sequence of Audio records with actor info
func (c *CorpusSearcher) buildSnippetCandidateWithActor(
	words []db.Audio,
	actor string,
	person string,
	targetBookId string,
	targetChapterNum int,
	targetVerseNum int,
) SnippetCandidate {
	if len(words) == 0 {
		return SnippetCandidate{}
	}

	first := words[0]
	last := words[len(words)-1]

	// Convert Audio records to Word records
	wordRecs := make([]db.Word, len(words))
	wordTexts := make([]string, len(words))
	for i, w := range words {
		wordRecs[i] = db.Word{
			WordId:      int(w.WordId),
			ScriptId:    int(w.ScriptId),
			WordSeq:     w.WordSeq,
			VerseNum:    w.VerseSeq,
			Word:        w.Text,
			WordBeginTS: w.BeginTS,
			WordEndTS:   w.EndTS,
			FAScore:     w.FAScore,
		}
		wordTexts[i] = w.Text
	}

	// Calculate distance
	distance := calculateDistance(
		first.BookId, first.ChapterNum, first.VerseSeq,
		targetBookId, targetChapterNum, targetVerseNum,
	)

	return SnippetCandidate{
		BookId:        first.BookId,
		ChapterNum:    first.ChapterNum,
		VerseStr:      first.VerseStr,
		ScriptId:      int(first.ScriptId),
		Actor:         actor,
		Person:        person,
		Words:         wordRecs,
		WordText:      strings.Join(wordTexts, " "),
		StartTS:       first.BeginTS,
		EndTS:         last.EndTS,
		MatchScore:    1.0, // Exact text match
		SpeakerMatch:  false, // TODO: Check actor/person against target
		PersonMatch:   false, // TODO: Check person against target
		ExactTextMatch: true,
		IsNearbyVerse: first.BookId == targetBookId && first.ChapterNum == targetChapterNum,
		Distance:      distance,
	}
}

// buildSnippetCandidate creates a SnippetCandidate from a sequence of Audio records (without actor info)
func (c *CorpusSearcher) buildSnippetCandidate(
	words []db.Audio,
	targetBookId string,
	targetChapterNum int,
	targetVerseNum int,
) SnippetCandidate {
	return c.buildSnippetCandidateWithActor(words, "", "", targetBookId, targetChapterNum, targetVerseNum)
}

// buildSnippetCandidateFromWord creates a SnippetCandidate from a single Word record
func (c *CorpusSearcher) buildSnippetCandidateFromWord(
	word db.Word,
	audioRec db.Audio,
	actor string,
	person string,
	targetBookId string,
	targetChapterNum int,
	targetVerseNum int,
) SnippetCandidate {
	distance := calculateDistance(
		audioRec.BookId, audioRec.ChapterNum, audioRec.VerseSeq,
		targetBookId, targetChapterNum, targetVerseNum,
	)

	return SnippetCandidate{
		BookId:        audioRec.BookId,
		ChapterNum:    audioRec.ChapterNum,
		VerseStr:      audioRec.VerseStr,
		ScriptId:      int(audioRec.ScriptId),
		Actor:         actor,
		Person:        person,
		Words:         []db.Word{word},
		WordText:      word.Word,
		StartTS:       word.WordBeginTS,
		EndTS:         word.WordEndTS,
		MatchScore:    0.8, // Single word match (lower than phrase)
		SpeakerMatch:  false, // TODO: Check against target
		PersonMatch:   false, // TODO: Check against target
		ExactTextMatch: true,
		IsNearbyVerse: audioRec.BookId == targetBookId && audioRec.ChapterNum == targetChapterNum,
		Distance:      distance,
	}
}

// resolveAudioFilePath resolves the audio file path relative to FCBH_DATASET_FILES
func (c *CorpusSearcher) resolveAudioFilePath(audioFile string) string {
	if filepath.IsAbs(audioFile) {
		return audioFile
	}
	return filepath.Join(c.basePath, audioFile)
}

// normalizeText normalizes text for matching (case-insensitive, trimmed)
func normalizeText(text string) string {
	return strings.TrimSpace(strings.ToLower(text))
}

// buildPhraseText builds a phrase from a sequence of Audio records
func buildPhraseText(words []db.Audio) string {
	texts := make([]string, len(words))
	for i, w := range words {
		texts[i] = w.Text
	}
	return strings.Join(texts, " ")
}

// calculateDistance calculates the distance between source and target location
// Returns a composite distance score (lower is closer)
func calculateDistance(
	sourceBookId string, sourceChapterNum int, sourceVerseNum int,
	targetBookId string, targetChapterNum int, targetVerseNum int,
) int {
	// Book distance: 0 if same book, 1000 if different (large penalty)
	bookDistance := 0
	if sourceBookId != targetBookId {
		bookDistance = 1000
	}

	// Chapter distance: absolute difference * 100
	chapterDistance := abs(sourceChapterNum-targetChapterNum) * 100

	// Verse distance: absolute difference
	verseDistance := abs(sourceVerseNum - targetVerseNum)

	// Total distance
	return bookDistance + chapterDistance + verseDistance
}

// abs returns absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// rankCandidates ranks candidates by match quality and distance
func rankCandidates(
	candidates []SnippetCandidate,
	targetBookId string,
	targetChapterNum int,
	targetVerseNum int,
) []SnippetCandidate {
	// Create a copy to avoid modifying original
	ranked := make([]SnippetCandidate, len(candidates))
	copy(ranked, candidates)

	// Calculate composite scores for ranking
	for i := range ranked {
		// Higher match score is better
		// Lower distance is better (invert for scoring)
		// Nearby verses get bonus
		score := ranked[i].MatchScore
		if ranked[i].IsNearbyVerse {
			score += 0.2 // Bonus for nearby verses
		}
		if ranked[i].ExactTextMatch {
			score += 0.1 // Bonus for exact text match
		}
		// Distance penalty (normalize distance to 0-1 range, then subtract)
		distancePenalty := float64(ranked[i].Distance) / 2000.0 // Max distance ~2000
		if distancePenalty > 1.0 {
			distancePenalty = 1.0
		}
		score -= distancePenalty * 0.3 // Max penalty of 0.3

		// Store in MatchScore for sorting
		ranked[i].MatchScore = score
	}

	// Sort by score (descending)
	for i := 0; i < len(ranked)-1; i++ {
		for j := i + 1; j < len(ranked); j++ {
			if ranked[i].MatchScore < ranked[j].MatchScore {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}

	return ranked
}

// GetResolvedAudioFilePath returns the resolved (absolute) path to the audio file for a snippet candidate
// This queries the database to get the audio_file from the scripts table
func (c *CorpusSearcher) GetResolvedAudioFilePath(candidate SnippetCandidate) (string, *log.Status) {
	if len(candidate.Words) == 0 {
		return "", log.ErrorNoErr(c.ctx, 400, "SnippetCandidate has no words")
	}

	// Query for the audio file path from the script
	query := `SELECT audio_file FROM scripts WHERE script_id = ?`
	var audioFile string
	err := c.conn.DB.QueryRow(query, candidate.ScriptId).Scan(&audioFile)
	if err != nil {
		return "", log.Error(c.ctx, 500, err, "Error querying audio file path")
	}

	return c.resolveAudioFilePath(audioFile), nil
}

