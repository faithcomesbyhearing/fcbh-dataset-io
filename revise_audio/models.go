package revise_audio

import (
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
)

// RevisionRequest represents a request to revise specific words/verses in audio
type RevisionRequest struct {
	DatasetName string
	Revisions   []Revision
}

// Revision specifies what needs to be changed in a verse
type Revision struct {
	BookId     string
	ChapterNum int
	VerseStr   string // e.g., "5", "6-10", "7a"
	
	// Words to replace: if empty, replace entire verse
	WordsToReplace []WordReplacement
	
	// Target speaker (Actor ID) - if empty, use same speaker as verse
	TargetActor string
	
	// Target Person (character) - fallback if Actor not found
	TargetPerson string
}

// WordReplacement specifies a word or phrase to replace
type WordReplacement struct {
	// Original word(s) to find and replace
	OriginalText string
	
	// Replacement text (what it should say)
	ReplacementText string
	
	// Word sequence indices in verse (if known)
	WordIndices []int
}

// SnippetCandidate represents a potential audio snippet from the corpus
type SnippetCandidate struct {
	// Source location
	BookId     string
	ChapterNum int
	VerseStr   string
	ScriptId   int
	
	// Speaker information
	Actor  string
	Person string
	
	// Word information
	Words      []db.Word // Word records with timestamps
	WordText   string    // Combined text of words
	StartTS    float64   // Start timestamp (first word)
	EndTS      float64   // End timestamp (last word)
	
	// Quality metrics
	MatchScore    float64 // How well it matches search criteria
	SpeakerMatch  bool    // Same Actor as target?
	PersonMatch   bool    // Same Person as target?
	ExactTextMatch bool   // Exact text match?
	
	// Context information
	IsNearbyVerse bool // Same book/chapter?
	Distance      int  // Verse distance from target (0 = same verse)
}

// VoiceConversionConfig holds configuration for voice conversion
type VoiceConversionConfig struct {
	// Tool selection
	UseRVC      bool // Use FCBH RVC system
	UseOpenVoice bool // Use OpenVoice (future)
	
	// RVC-specific settings
	RVCModelPath      string // Path to trained RVC model
	SpeakerEmbeddingPath string // Path to speaker embedding file
	F0Method          string // "rmvpe", "harvest", "dio", "pm"
	
	// Output settings
	SampleRate int // Output sample rate (default: 16000)
}

// ProsodyConfig holds configuration for prosody matching
type ProsodyConfig struct {
	// Tool selection
	UseDSP      bool // Use DSP-based prosody matching
	UseNeural   bool // Use neural prosody transfer (future)
	
	// DSP-specific settings
	F0Method        string // "pyworld", "librosa", "rmvpe"
	PitchShiftRange float64 // Max semitones to shift (default: 2.0)
	TimeStretchRange float64 // Max time stretch factor (default: 1.2)
	
	// Reference audio for prosody extraction
	ReferenceContextSeconds float64 // Seconds before/after to analyze (default: 5.0)
}

// AudioRevisionResult represents the result of revising audio
type AudioRevisionResult struct {
	// Original information
	OriginalVerse db.Script
	OriginalWords []db.Word
	
	// Revision information
	Revision Revision
	
	// Source of replacement
	SourceSnippet SnippetCandidate
	
	// Processing steps
	VoiceConverted bool
	ProsodyMatched bool
	
	// Output
	RevisedAudioPath string // Path to revised audio snippet
	StartTS          float64 // Start time in chapter
	EndTS            float64 // End time in chapter
	
	// Quality metrics
	ProcessingTimeMs int64
	QualityScore     float64 // Subjective quality score (0-1)
}

// ChapterRevisionResult represents the complete result of revising a chapter
type ChapterRevisionResult struct {
	BookId     string
	ChapterNum int
	AudioFile  string // Original chapter audio file
	
	// Revisions applied
	Revisions []AudioRevisionResult
	
	// Output
	RevisedAudioPath string // Path to revised chapter audio
	OriginalBackupPath string // Path to backup of original
	
	// Statistics
	TotalRevisions int
	WordsReplaced  int
	ProcessingTimeMs int64
}

