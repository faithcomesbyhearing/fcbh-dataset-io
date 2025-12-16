package revise_audio

import (
	"context"
	"os"

	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

// SegmentBoundary represents a detected boundary point in audio
type SegmentBoundary struct {
	Time      float64 // Time in seconds
	Confidence float64 // Confidence score (0-1, higher = more confident it's a good cut point)
	Type      string  // Type: "silence", "pause", "word_boundary"
}

// SegmentBoundaryDetector finds good cut-points in audio for clean segment replacement
// Uses high-confidence word boundaries from the database
type SegmentBoundaryDetector struct {
	ctx     context.Context
	tempDir string
}

// NewSegmentBoundaryDetector creates a new boundary detector
func NewSegmentBoundaryDetector(ctx context.Context) *SegmentBoundaryDetector {
	return &SegmentBoundaryDetector{
		ctx: ctx,
	}
}

// DetectBoundariesFromDB finds good segment boundaries using high-confidence word timestamps from database
// 
// GENERALIZED RULES FOR SEGMENT BOUNDARIES (derived from data analysis):
// 1. LARGE GAP (>300ms) between words = natural pause = excellent boundary
//    - Gap >500ms: +3.0 points (very large pause, likely sentence/phrase boundary)
//    - Gap >300ms: +2.0 points (large pause, likely phrase boundary)
//    - Gap >150ms: +1.0 points (noticeable pause)
// 2. HIGH FA SCORE (>=0.9) on both sides = reliable word timings = good boundary
//    - FA >=0.95 on either side: +1.0 point
//    - FA >=0.9 on either side: +0.5 points
//    - Both sides FA >=0.9: +0.5 bonus points
// 3. COMBINATION: Large gap + high FA on both sides = BEST boundary (score >= 4.0)
// 4. End of verse/chapter = natural boundary (always use)
// 5. Start of verse/chapter = natural boundary (always use)
//
// Boundary score >= 2.0 = good boundary for segment cuts
// Confidence is normalized to 0-1 range (score / 5.5 max)
//
// Returns list of boundaries sorted by time with confidence scores
func (d *SegmentBoundaryDetector) DetectBoundariesFromDB(
	words []struct {
		BeginTS float64
		EndTS   float64
		FAScore float64
	},
	minConfidence float64,
) []SegmentBoundary {
	if minConfidence <= 0 {
		minConfidence = 0.8 // Default 80% confidence
	}

	boundaries := []SegmentBoundary{}
	
	if len(words) == 0 {
		return boundaries
	}
	
	// Analyze gaps between words to find natural boundaries
	// Rule: Large gaps (>300ms) + high FA scores on both sides = excellent boundaries
	for i := 0; i < len(words)-1; i++ {
		currentWord := words[i]
		nextWord := words[i+1]
		
		// Calculate gap between words
		gap := nextWord.BeginTS - currentWord.EndTS
		
		// Calculate boundary quality score
		// Higher score = better boundary for cutting
		score := 0.0
		boundaryType := "word_boundary"
		
		// Gap scoring (larger gaps = better boundaries)
		if gap > 0.5 { // 500ms+ = excellent boundary
			score += 3.0
			boundaryType = "large_pause"
		} else if gap > 0.3 { // 300ms+ = very good boundary
			score += 2.0
			boundaryType = "pause"
		} else if gap > 0.15 { // 150ms+ = good boundary
			score += 1.0
			boundaryType = "small_pause"
		}
		
		// FA score scoring (higher confidence = more reliable)
		if currentWord.FAScore >= 0.95 {
			score += 1.0
		} else if currentWord.FAScore >= 0.9 {
			score += 0.5
		}
		
		if nextWord.FAScore >= 0.95 {
			score += 1.0
		} else if nextWord.FAScore >= 0.9 {
			score += 0.5
		}
		
		// Bonus for both sides having high FA
		if currentWord.FAScore >= 0.9 && nextWord.FAScore >= 0.9 {
			score += 0.5
		}
		
		// Normalize score to 0-1 range for confidence
		// Max possible score: 3.0 (gap) + 1.0 (FA before) + 1.0 (FA after) + 0.5 (both) = 5.5
		confidence := score / 5.5
		if confidence > 1.0 {
			confidence = 1.0
		}
		
		// Only add boundaries with meaningful confidence
		// Include boundaries with either:
		// 1. Large gap (>150ms) - natural pause
		// 2. High FA on both sides (>=0.9) - reliable timings
		// 3. Combination of both (best)
		if gap > 0.15 || (currentWord.FAScore >= 0.9 && nextWord.FAScore >= 0.9) {
			boundaries = append(boundaries, SegmentBoundary{
				Time:      currentWord.EndTS, // Boundary is at end of current word
				Confidence: confidence,
				Type:      boundaryType,
			})
		}
	}
	
	// Also add word boundaries with high confidence (for cases without large gaps)
	for _, word := range words {
		if word.FAScore >= minConfidence {
			// Add start boundary (only if not already added as end of previous word)
			boundaries = append(boundaries, SegmentBoundary{
				Time:      word.BeginTS,
				Confidence: word.FAScore,
				Type:      "word_start",
			})
			// Add end boundary (may duplicate with gap-based boundaries, but that's OK)
			boundaries = append(boundaries, SegmentBoundary{
				Time:      word.EndTS,
				Confidence: word.FAScore,
				Type:      "word_end",
			})
		}
	}

	// Sort boundaries by time
	for i := 0; i < len(boundaries)-1; i++ {
		for j := i + 1; j < len(boundaries); j++ {
			if boundaries[i].Time > boundaries[j].Time {
				boundaries[i], boundaries[j] = boundaries[j], boundaries[i]
			}
		}
	}

	return boundaries
}

// FindBestBoundaries finds the best boundaries around a target time range
// Returns start and end boundaries that best encompass the target range
func (d *SegmentBoundaryDetector) FindBestBoundaries(
	boundaries []SegmentBoundary,
	targetStart float64,
	targetEnd float64,
) (float64, float64, *log.Status) {
	if len(boundaries) == 0 {
		// No boundaries found, use target times as-is
		return targetStart, targetEnd, nil
	}

	// Find closest boundary before targetStart
	bestStart := targetStart
	bestStartConfidence := 0.0
	for _, b := range boundaries {
		if b.Time <= targetStart && b.Time >= targetStart-0.5 { // Within 500ms before
			if b.Confidence > bestStartConfidence {
				bestStart = b.Time
				bestStartConfidence = b.Confidence
			}
		}
	}

	// Find closest boundary after targetEnd
	bestEnd := targetEnd
	bestEndConfidence := 0.0
	for _, b := range boundaries {
		if b.Time >= targetEnd && b.Time <= targetEnd+0.5 { // Within 500ms after
			if b.Confidence > bestEndConfidence {
				bestEnd = b.Time
				bestEndConfidence = b.Confidence
			}
		}
	}

	return bestStart, bestEnd, nil
}

// Cleanup removes temporary files
func (d *SegmentBoundaryDetector) Cleanup() {
	if d.tempDir != "" {
		os.RemoveAll(d.tempDir)
		d.tempDir = ""
	}
}

