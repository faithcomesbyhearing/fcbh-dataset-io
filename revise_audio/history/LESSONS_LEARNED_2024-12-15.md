# Lessons Learned - Audio Revision Pipeline Evaluation
**Date**: 2024-12-15  
**Session**: TTS + VITS + Prosody Matching Quality Evaluation

## Executive Summary

The complete audio revision pipeline (MMS TTS → VITS Voice Conversion → Prosody Matching) was evaluated and found to produce **unacceptable quality** despite acceptable TTS base. Root causes: insufficient training data, cascading quality loss, and tool limitations.

## Test Results Summary

| Component | Status | Quality | Notes |
|-----------|--------|---------|-------|
| MMS TTS | ✅ | Acceptable | User confirmed "the segment alone is ok" |
| VITS Voice Conversion | ❌ | Unacceptable | Artifacts, quality degradation |
| Prosody Matching | ❌ | Unacceptable | DSP artifacts, no quality improvement |
| Gap Preservation | ✅ | Working | Correctly preserves 0.700s gap |

## Critical Issues Identified

### 1. Insufficient Training Data
- **Current**: 26 verse segments (~4 minutes)
- **Required**: 30+ minutes of clean, diverse speech
- **Impact**: Model overfits, fails to generalize, produces artifacts
- **Solution**: Collect more training data before retraining

### 2. Cascading Quality Loss
- **Problem**: Each processing step compounds errors
  - TTS → introduces artifacts
  - Voice conversion → adds more artifacts
  - Prosody matching → introduces DSP artifacts
- **Impact**: Final quality poor despite acceptable TTS base
- **Solution**: Simplify pipeline or use end-to-end training

### 3. Tool Limitations
- **VITS**: Requires more training data than available
- **Prosody Matching**: DSP-based approach too simplistic
- **Impact**: Cannot achieve target quality with current tools/data
- **Solution**: Consider alternatives or collect more data

## What Works

1. ✅ **MMS TTS**: Produces acceptable quality for text-to-speech
2. ✅ **Gap Preservation**: Logic correctly preserves silence gaps
3. ✅ **Segment Boundary Detection**: Identifies good cut points
4. ✅ **Audio Stitching**: Cross-fade and concatenation work correctly

## What Doesn't Work

1. ❌ **VITS Voice Conversion**: Quality unacceptable with limited data
2. ❌ **Prosody Matching**: DSP-based approach introduces artifacts
3. ❌ **Cascaded Pipeline**: Multiple steps compound quality issues

## Recommendations

### Immediate (Short Term)
1. **Use Corpus Snippets**: When available, use actual audio from corpus
   - Zero quality loss
   - Natural prosody
   - No training needed
2. **Simplify Pipeline**: Use TTS only (no voice conversion)
   - Accept voice mismatch
   - Fewer artifacts
   - Better intelligibility

### Next Steps (Medium Term)
1. **Collect More Training Data**: 30+ minutes of clean, diverse speech
2. **Evaluate Alternative TTS**: Consider speaker-adaptive TTS (XTTS, Coqui)
3. **Retrain VITS**: After collecting sufficient data

### Future (Long Term)
1. **Neural Prosody Matching**: Replace DSP with learned prosody transfer
2. **End-to-End Training**: Single model for TTS + voice matching
3. **Quality Metrics**: Establish objective quality measurements

## Beads Issues to Create

1. **Evaluate TTS + VITS + Prosody pipeline quality** (Priority: 1)
   - Status: Complete (evaluation done)
   - Findings documented in SESSION_SUMMARY_2024-12-15_TTS_VITS_PROSOODY_EVALUATION.md

2. **Implement gap preservation for audio stitching** (Priority: 2)
   - Status: Done
   - Function: ReplaceSegmentInChapterWithGapMatching()

3. **Collect more VITS training data** (Priority: 1)
   - Status: Open
   - Need: 30+ minutes of clean, diverse speech

4. **Evaluate corpus-based approach** (Priority: 1)
   - Status: Open
   - Use corpus snippets when available, fall back to TTS

## Files Reference

- **Session Summary**: `history/SESSION_SUMMARY_2024-12-15_TTS_VITS_PROSOODY_EVALUATION.md`
- **Test Outputs**: `/home/ec2-user/tmp/arti_revised_audio/`
  - `JUD_01_v1_raw_tts_segment.mp3` - Raw TTS (acceptable)
  - `JUD_01_v1_tts_voice_converted.mp3` - TTS + VC (unacceptable)
  - `JUD_01_v1_tts_prosody_matched.mp3` - TTS + Prosody (unacceptable)
  - `JUD_01_v1_tts_only.mp3` - Full verse with TTS replacement

## Conclusion

The current approach is **too ambitious for available training data**. Quality issues are fundamental and require either:
1. More training data (30+ minutes)
2. Simpler pipeline (TTS only)
3. Alternative approach (corpus snippets)

The infrastructure (gap preservation, stitching) works well and can be reused regardless of audio generation approach.

