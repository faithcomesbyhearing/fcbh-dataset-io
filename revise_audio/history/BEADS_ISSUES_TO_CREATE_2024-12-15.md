# Beads Issues to Create - 2024-12-15

## Summary
After evaluating the TTS + VITS + Prosody pipeline, the following issues should be created in beads:

## Issues to Create

### 1. Evaluate TTS + VITS + Prosody Pipeline Quality
**Type**: task  
**Priority**: 1  
**Status**: done  
**Body**:
```
Tested complete pipeline: MMS TTS → VITS voice conversion → prosody matching.

Results:
- TTS: Acceptable quality (user confirmed "the segment alone is ok")
- Voice Conversion: Unacceptable quality (artifacts, quality degradation)
- Prosody Matching: Unacceptable quality (DSP artifacts, no improvement)

Root causes:
- Insufficient training data (26 segments vs 30+ min needed)
- Cascading quality loss through multiple processing steps
- DSP prosody matching limitations

See: history/SESSION_SUMMARY_2024-12-15_TTS_VITS_PROSOODY_EVALUATION.md
```

### 2. Implement Gap Preservation for Audio Stitching
**Type**: task  
**Priority**: 2  
**Status**: done  
**Body**:
```
Implemented step-by-step gap matching approach:
1. Measure TTS leading silence
2. Measure original gap at boundary
3. Place TTS appropriately (with/without cross-fade based on silence)

Function: ReplaceSegmentInChapterWithGapMatching() in audio_stitch.go
Test: test_jude_tts_only/main.go

Status: Complete and tested successfully
Result: Correctly preserves 0.700s gap at boundary
```

### 3. Collect More VITS Training Data
**Type**: task  
**Priority**: 1  
**Status**: open  
**Body**:
```
Current training data (26 segments, ~4 minutes) is insufficient for quality voice conversion.

Requirements:
- Need 30+ minutes of clean, diverse speech
- Good phonetic coverage
- Consistent recording quality

Impact: This is blocking acceptable voice conversion quality.

Recommendation: Collect more data before retraining VITS model.
```

### 4. Evaluate Corpus-Based Approach for Audio Revision
**Type**: task  
**Priority**: 1  
**Status**: open  
**Body**:
```
Alternative to TTS+VC pipeline: use actual audio snippets from corpus when available.

Pros:
- Zero quality loss (real human speech)
- Natural prosody (already matches context)
- No training needed (works immediately)

Cons:
- Requires good corpus coverage
- May not always have matching snippets

Recommendation: Use as primary approach, fall back to TTS only when corpus has no matches.

See: history/VOICE_CONVERSION_ALTERNATIVES.md
```

### 5. Simplify Audio Revision Pipeline
**Type**: task  
**Priority**: 1  
**Status**: open  
**Body**:
```
Current pipeline (TTS → VC → Prosody) produces unacceptable quality due to cascading errors.

Simplified approach:
- Use TTS only (no voice conversion)
- Accept that voice won't match perfectly
- Fewer artifacts, better intelligibility

Alternative: Use corpus snippets when available (zero quality loss).

Status: Recommendation pending decision
```

## Commands to Create Issues

```bash
# Issue 1: Pipeline evaluation (done)
bd create "Evaluate TTS + VITS + Prosody pipeline quality" \
  --body "Tested complete pipeline: MMS TTS → VITS voice conversion → prosody matching. Results: TTS acceptable, voice conversion and prosody matching both produced unacceptable quality. Root causes: insufficient training data (26 segments vs 30+ min needed), cascading quality loss, DSP prosody limitations. See history/SESSION_SUMMARY_2024-12-15_TTS_VITS_PROSOODY_EVALUATION.md" \
  --type task --priority 1 --status done

# Issue 2: Gap preservation (done)
bd create "Implement gap preservation for audio stitching" \
  --body "Implemented step-by-step gap matching: measure TTS leading silence, measure original gap, place TTS appropriately (with/without cross-fade). Function: ReplaceSegmentInChapterWithGapMatching(). Tested successfully - preserves 0.700s gap at boundary. Status: Complete." \
  --type task --priority 2 --status done

# Issue 3: Collect training data
bd create "Collect more VITS training data for quality improvement" \
  --body "Current training data (26 segments, ~4 min) insufficient for quality voice conversion. Need 30+ minutes of clean, diverse speech with good phonetic coverage. This is blocking acceptable voice conversion quality. Recommendation: collect more data before retraining VITS model." \
  --type task --priority 1

# Issue 4: Corpus-based approach
bd create "Evaluate corpus-based approach for audio revision" \
  --body "Alternative to TTS+VC: use actual audio snippets from corpus when available. Pros: zero quality loss, natural prosody, no training needed. Cons: requires good corpus coverage. Recommendation: use as primary approach, fall back to TTS only when corpus has no matches." \
  --type task --priority 1

# Issue 5: Simplify pipeline
bd create "Simplify audio revision pipeline" \
  --body "Current pipeline (TTS → VC → Prosody) produces unacceptable quality due to cascading errors. Simplified approach: Use TTS only (no voice conversion), accept voice mismatch, fewer artifacts. Alternative: Use corpus snippets when available. Status: Recommendation pending decision." \
  --type task --priority 1
```

## Related Files

- `history/SESSION_SUMMARY_2024-12-15_TTS_VITS_PROSOODY_EVALUATION.md` - Full session summary
- `history/LESSONS_LEARNED_2024-12-15.md` - Lessons learned summary
- `history/VOICE_CONVERSION_ALTERNATIVES.md` - Alternative approaches
- Test outputs: `/home/ec2-user/tmp/arti_revised_audio/`

