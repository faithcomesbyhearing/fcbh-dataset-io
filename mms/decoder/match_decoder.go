package decoder

import (
	"container/heap"
)

// Hypothesis represents a partial decoding hypothesis
type Hypothesis struct {
	// Sequence of tokens/words decoded so far
	Tokens []int

	// Position in the expected output sequence
	ExpectedPos int

	// Cumulative log probability score
	Score float64

	// Position in acoustic frames
	TimeStep int

	// For tracking beam search priority
	index int
}

// Define beam priority queue methods (for heap interface)
type BeamPQ []*Hypothesis

func (pq BeamPQ) Len() int           { return len(pq) }
func (pq BeamPQ) Less(i, j int) bool { return pq[i].Score > pq[j].Score } // Higher score = higher priority
func (pq BeamPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *BeamPQ) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Hypothesis)
	item.index = n
	*pq = append(*pq, item)
}
func (pq *BeamPQ) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// ExpectedMatchingBeamDecoder implements a beam search decoder with expected matching
type ExpectedMatchingBeamDecoder struct {
	// Expected sequence of words/tokens
	ExpectedSequence []int

	// Vocabulary size of model
	VocabSize int

	// Size of beam
	BeamWidth int

	// CTC blank token ID
	BlankID int

	// Weights for scoring
	ExpectedMatchBonus float64 // Bonus for matching expected token
	InsertionPenalty   float64 // Penalty for inserting a token not in expected sequence
	DeletionPenalty    float64 // Penalty for skipping an expected token

	// Maximum lookahead in expected sequence
	MaxLookahead int
}

// Decode performs beam search decoding on acoustic model probabilities
func (d *ExpectedMatchingBeamDecoder) Decode(logProbs [][]float64) []int {
	numFrames := len(logProbs)

	// Initialize beam with empty hypothesis
	beam := &BeamPQ{}
	heap.Init(beam)

	// Start with single empty hypothesis
	heap.Push(beam, &Hypothesis{
		Tokens:      []int{},
		ExpectedPos: 0,
		Score:       0.0,
		TimeStep:    0,
	})

	// Perform beam search through all time steps
	for t := 0; t < numFrames; t++ {
		currBeam := &BeamPQ{}
		heap.Init(currBeam)

		// Process current beam of hypotheses
		for beam.Len() > 0 && currBeam.Len() < d.BeamWidth {
			hyp := heap.Pop(beam).(*Hypothesis)

			// Skip if this hypothesis is from a future time step
			if hyp.TimeStep > t {
				heap.Push(currBeam, hyp)
				continue
			}

			// Get logProbs for current time step
			frameLogProbs := logProbs[t]

			// Consider emitting each possible token
			for tokID := 0; tokID < d.VocabSize; tokID++ {
				// Skip blank token in output (CTC-specific logic)
				if tokID == d.BlankID {
					// Just advance time step for blank, keeping same hypothesis
					newHyp := &Hypothesis{
						Tokens:      append([]int{}, hyp.Tokens...),
						ExpectedPos: hyp.ExpectedPos,
						Score:       hyp.Score + frameLogProbs[tokID],
						TimeStep:    t + 1,
					}
					heap.Push(currBeam, newHyp)
					continue
				}

				// 1. Emit token - calculate new score with expected output considerations
				var matchBonus float64 = 0

				// Check if this token matches current expected token
				if hyp.ExpectedPos < len(d.ExpectedSequence) &&
					tokID == d.ExpectedSequence[hyp.ExpectedPos] {
					matchBonus = d.ExpectedMatchBonus
				} else {
					// Look ahead in expected sequence for potential match
					foundMatch := false
					for i := 1; i <= d.MaxLookahead && hyp.ExpectedPos+i < len(d.ExpectedSequence); i++ {
						if tokID == d.ExpectedSequence[hyp.ExpectedPos+i] {
							// Found match in lookahead window, but apply deletion penalty for skipped tokens
							matchBonus = d.ExpectedMatchBonus - float64(i)*d.DeletionPenalty
							foundMatch = true
							break
						}
					}

					// Apply insertion penalty if not found in lookahead
					if !foundMatch {
						matchBonus = d.InsertionPenalty
					}
				}

				// Calculate new hypothesis score
				newScore := hyp.Score + frameLogProbs[tokID] + matchBonus

				// Determine new expected position
				newExpectedPos := hyp.ExpectedPos
				// If token matched current expected token, advance expected position
				if hyp.ExpectedPos < len(d.ExpectedSequence) &&
					tokID == d.ExpectedSequence[hyp.ExpectedPos] {
					newExpectedPos++
				} else {
					// Check if token matched something in lookahead window
					for i := 1; i <= d.MaxLookahead && hyp.ExpectedPos+i < len(d.ExpectedSequence); i++ {
						if tokID == d.ExpectedSequence[hyp.ExpectedPos+i] {
							// Advance past the matched position
							newExpectedPos = hyp.ExpectedPos + i + 1
							break
						}
					}
				}

				// Create new hypothesis with this token
				newTokens := append(append([]int{}, hyp.Tokens...), tokID)
				newHyp := &Hypothesis{
					Tokens:      newTokens,
					ExpectedPos: newExpectedPos,
					Score:       newScore,
					TimeStep:    t + 1,
				}

				heap.Push(currBeam, newHyp)
			}
		}

		// Replace beam with current hypotheses for next time step
		beam = currBeam
	}

	// Return best hypothesis
	if beam.Len() > 0 {
		bestHyp := heap.Pop(beam).(*Hypothesis)
		return bestHyp.Tokens
	}

	return []int{} // Empty result if beam is empty
}
