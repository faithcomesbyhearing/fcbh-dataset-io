# Audio Revision System - Session Summary
**Date**: 2024-12-15  
**Focus**: TTS + Voice Conversion + Prosody Matching Evaluation

## Session Overview

This session evaluated the complete audio revision pipeline: MMS TTS → VITS Voice Conversion → Prosody Matching. The goal was to revise Jude 1:1 by replacing the second half (from "who are loved..." onwards) with TTS-generated audio that matches the original narrator's voice and prosody.

## Test Results

### 1. MMS TTS Only ✅
- **Test**: Generate TTS for "who are loved in God the Father and kept for Jesus Christ:"
- **Result**: Acceptable quality - user confirmed "the segment alone is ok"
- **Output**: `/home/ec2-user/tmp/arti_revised_audio/JUD_01_v1_raw_tts_segment.mp3`
- **Finding**: MMS TTS produces intelligible output suitable as a base

### 2. VITS Voice Conversion ❌
- **Test**: Apply voice conversion to TTS segment using G_13600.pth checkpoint
- **Model**: So-VITS-SVC trained on 26 verse segments from Jude chapter 1
- **Result**: **Unacceptable quality** - user did not like the result
- **Output**: `/home/ec2-user/tmp/arti_revised_audio/JUD_01_v1_tts_voice_converted.mp3`
- **Finding**: Voice conversion introduces artifacts and quality degradation

### 3. Prosody Matching ❌
- **Test**: Apply prosody matching to TTS segment using original audio as reference
- **Reference**: Original audio segment from "who" (9.620s) to end of verse 1 (12.970s)
- **Method**: DSP-based (pitch shift, time stretch, energy matching)
- **Result**: **Unacceptable quality** - user did not like the result
- **Output**: `/home/ec2-user/tmp/arti_revised_audio/JUD_01_v1_tts_prosody_matched.mp3`
- **Finding**: Prosody matching introduces artifacts and doesn't improve quality

## Key Findings

### 1. Limited Training Data
- **Issue**: Only 26 verse segments (~4 minutes) used for VITS training
- **Impact**: Voice conversion models typically need 30+ minutes of diverse, clean speech
- **Result**: Model likely overfits to specific phrases, fails to generalize, produces artifacts

### 2. Cascading Quality Loss
- **Problem**: Each processing step compounds errors:
  - MMS TTS → introduces TTS artifacts
  - VITS voice conversion → adds conversion artifacts  
  - Prosody matching → introduces DSP artifacts
- **Result**: Final output quality is poor despite acceptable TTS base

### 3. VITS Model Limitations
- **Training Status**: Model at epoch 13600, but with limited data, more epochs won't help
- **Requirements**: So-VITS-SVC works best with:
  - 30+ minutes of training data
  - Diverse phonetic coverage
  - Clean, consistent recordings
- **Current State**: Training data insufficient for quality voice conversion

### 4. Prosody Matching Limitations
- **Method**: DSP-based (pitch shift, time stretch, energy matching)
- **Issues**:
  - Too simplistic to capture nuanced prosodic patterns
  - Can introduce phase distortion and robotic artifacts
  - Doesn't account for natural speech rhythm variations
- **Result**: Doesn't meaningfully improve quality

### 5. MMS TTS as Base
- **Finding**: MMS TTS quality is acceptable on its own
- **Issue**: May not be optimal foundation for voice conversion
- **Note**: Multilingual but not optimized for voice quality

## Gap Preservation Implementation

### Approach
Implemented step-by-step gap matching:
1. Generate TTS segment
2. Measure leading silence in TTS
3. Measure original gap at boundary
4. Place TTS appropriately:
   - If no leading silence: place at end of gap (no cross-fade)
   - If leading silence exists: place with cross-fade to match original gap

### Implementation
- **Function**: `ReplaceSegmentInChapterWithGapMatching()` in `audio_stitch.go`
- **Test**: `test_jude_tts_only/main.go`
- **Status**: ✅ Implemented and tested
- **Result**: Gap preservation logic works correctly (0.700s gap preserved)

## Lessons Learned

### What Works
1. ✅ **MMS TTS**: Produces acceptable quality for text-to-speech
2. ✅ **Gap Preservation**: Logic correctly preserves silence gaps at boundaries
3. ✅ **Segment Boundary Detection**: Successfully identifies good cut points
4. ✅ **Audio Stitching**: Cross-fade and concatenation work correctly

### What Doesn't Work
1. ❌ **VITS Voice Conversion**: Quality unacceptable with limited training data
2. ❌ **Prosody Matching**: DSP-based approach introduces artifacts
3. ❌ **Cascaded Pipeline**: TTS → VC → Prosody compounds quality issues

### Root Causes
1. **Insufficient Training Data**: 26 segments (~4 min) is far below recommended 30+ minutes
2. **Over-Complexity**: Too many processing steps compound errors
3. **Tool Limitations**: DSP-based prosody matching too simplistic
4. **Model Mismatch**: So-VITS-SVC may not be optimal for this use case

## Recommendations

### Short Term (Immediate)
1. **Use Corpus Snippets**: When available, use actual audio snippets from corpus
   - Zero quality loss
   - Natural prosody
   - No training needed
   - Fall back to TTS only when corpus has no matches

2. **Simplify Pipeline**: Use TTS only (no voice conversion)
   - Accept that voice won't match perfectly
   - Fewer artifacts
   - Better intelligibility

### Medium Term (Next Steps)
1. **Collect More Training Data**: 
   - Target: 30+ minutes of clean, diverse speech
   - Ensure phonetic coverage
   - Retrain VITS model

2. **Evaluate Alternative TTS**: 
   - Consider speaker-adaptive TTS (XTTS, Coqui TTS)
   - Fine-tune on narrator's voice
   - Skip voice conversion entirely

### Long Term (Future)
1. **Neural Prosody Matching**: 
   - Replace DSP-based approach with learned prosody transfer
   - Extract prosody embeddings from reference audio
   - Condition decoder on prosody

2. **End-to-End Training**: 
   - Train single model for TTS + voice matching
   - Avoid cascading quality loss
   - Better quality control

## Files Created/Modified

### Test Scripts
- `revise_audio/cmd/test_jude_tts_only/main.go` - TTS-only test with gap preservation
- `revise_audio/cmd/test_voice_convert_segment/main.go` - Voice conversion test
- `revise_audio/cmd/test_prosody_match_segment/main.go` - Prosody matching test

### Core Implementation
- `revise_audio/audio_stitch.go` - Added `ReplaceSegmentInChapterWithGapMatching()`
- `revise_audio/audio_stitch.go` - Added `measureLeadingSilence()` helper

### Output Files
- `/home/ec2-user/tmp/arti_revised_audio/JUD_01_v1_raw_tts_segment.mp3` - Raw TTS
- `/home/ec2-user/tmp/arti_revised_audio/JUD_01_v1_tts_voice_converted.mp3` - TTS + VC
- `/home/ec2-user/tmp/arti_revised_audio/JUD_01_v1_tts_prosody_matched.mp3` - TTS + Prosody
- `/home/ec2-user/tmp/arti_revised_audio/JUD_01_v1_tts_only.mp3` - Full verse with TTS replacement

## Conclusion

The current approach (TTS → Voice Conversion → Prosody Matching) is **too ambitious for the available training data**. Quality issues stem from:

1. Insufficient training data (26 segments vs. recommended 30+ minutes)
2. Cascading quality loss through multiple processing steps
3. Limitations of DSP-based prosody matching

**Recommendation**: Pivot to simpler approach:
- Use corpus snippets when available (zero quality loss)
- Use TTS only when corpus has no matches (acceptable quality)
- Collect more training data before attempting voice conversion
- Consider alternative TTS systems optimized for voice quality

The gap preservation and stitching infrastructure works well and can be reused regardless of the audio generation approach.

