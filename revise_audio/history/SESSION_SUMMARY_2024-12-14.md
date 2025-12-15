# Audio Revision System - Session Summary
**Date**: 2024-12-14  
**Focus**: Voice Conversion Implementation and Evaluation

## Session Overview

This session focused on implementing and evaluating voice conversion for the audio revision system. We progressed from initial TTS experiments through voice conversion attempts, ultimately determining that neural modeling (RVC) is required for acceptable quality.

---

## Key Accomplishments

### 1. TTS Integration ✅
- Implemented MMS TTS adapter for generating replacement audio
- Tested single-word, phrase, and segment-level generation
- **Finding**: Segment-level TTS with high-confidence boundaries produces best results

### 2. Segment Boundary Detection ✅
- Implemented `SegmentBoundaryDetector` with scoring system
- Uses gap duration + FA scores to identify optimal cut points
- **Generalized Rules**:
  - Large gaps (>300ms) = natural pauses = strong boundaries
  - High FA scores (>=0.9) on both sides = reliable timings
  - Score >= 2.0 = good boundary, >= 4.0 = excellent

### 3. Voice Conversion Attempts ⚠️
- Implemented DSP-based voice conversion (pitch shift + loudness normalization)
- Attempted spectral envelope transfer (abandoned due to quality issues)
- **Finding**: DSP approaches produce intelligible but unacceptable results
- **Conclusion**: Neural modeling (RVC) required for acceptable quality

### 4. Prosody Matching ✅
- Implemented DSP-based prosody matching (F0, energy, timing)
- Uses pyworld/librosa for pitch extraction
- Temporarily disabled to focus on voice conversion quality

---

## Key Learnings

### Voice Conversion Quality Issues

1. **Simple Pitch Shifting + Loudness**: 
   - Intelligible but sounds robotic
   - Cannot convert TTS to natural human speech

2. **Spectral Envelope Transfer**:
   - Attempted using Griffin-Lim reconstruction
   - Result: Completely unintelligible
   - Griffin-Lim introduces too many artifacts

3. **Conservative Approach** (Final Attempt):
   - Pitch shift + loudness normalization only
   - Uses exact segment from original as reference
   - Result: Intelligible but not acceptable quality

### Corpus Snippets vs TTS

- **Corpus Snippets**: Tried initially, but word boundaries too imprecise
- **TTS Approach**: Needed for generating new content
- **Segment Boundaries**: Valuable building block for precise cuts

### MMS TTS Limitations

- **No Speaker Embedding Support**: Base MMS TTS models don't accept external speaker embeddings
- **Multi-speaker Versions**: Exist but require fine-tuning with multi-speaker datasets
- **Conclusion**: Must use generate-then-convert approach

---

## Decision: So-VITS-SVC (Option 4)

After evaluating alternatives, we decided to proceed with **So-VITS-SVC** for voice conversion.

### Why So-VITS-SVC?

1. **Mature and Battle-Tested**: Well-established open-source project
2. **Better Documentation**: Easier integration than fixing FCBH RVC
3. **Community Support**: Active community for troubleshooting
4. **Modern Techniques**: Uses HuBERT, RMVPE (same as FCBH RVC)
5. **Faster Path to Working**: More likely to work out of the box

### Why Not Other Options?

- **Option 1 (Corpus Snippets)**: Word boundaries too imprecise
- **Option 2 (Formant Shifting)**: Still DSP-based, insufficient
- **Option 3 (FCBH RVC)**: Has compatibility issues, less mature
- **Option 5 (OpenVoice/XTTS-v2)**: XTTS limited languages, OpenVoice less mature

---

## Current System State

### Working Components ✅
- TTS generation (MMS TTS)
- Segment boundary detection
- Audio stitching with cross-fades
- Prosody matching (implemented, temporarily disabled)
- Voice conversion infrastructure (Go-Python integration)

### Needs Work ⚠️
- Voice conversion quality (requires So-VITS-SVC integration)
- Speaker embedding extraction (currently simple fallback)

### Next Steps
1. Research So-VITS-SVC integration approach
2. Install and configure So-VITS-SVC
3. Train model on speaker corpus data
4. Integrate inference into `revise_audio` pipeline
5. Re-enable prosody matching after VC quality is acceptable

---

## Files Created/Modified

### New Files
- `revise_audio/python/voice_conversion.py` - Voice conversion Python module
- `revise_audio/python/prosody_match.py` - Prosody matching Python module
- `revise_audio/vc_adapter.go` - Voice conversion Go adapter
- `revise_audio/prosody_adapter.go` - Prosody matching Go adapter
- `revise_audio/segment_boundary.go` - Segment boundary detection
- `revise_audio/cmd/test_tts_segments/main.go` - Test script for full workflow
- `revise_audio/cmd/test_vc/main.go` - Voice conversion test
- `revise_audio/cmd/test_prosody/main.go` - Prosody matching test
- `revise_audio/history/VOICE_CONVERSION_ALTERNATIVES.md` - Alternatives analysis
- `revise_audio/history/RVC_EXPLANATION.md` - RVC concepts explained

### Modified Files
- `revise_audio/models.go` - Added VoiceConversionConfig, ProsodyConfig
- `revise_audio/mms_tts_adapter.go` - MMS TTS integration
- `revise_audio/audio_stitch.go` - Audio stitching with cross-fades
- `revise_audio/snippet_extract.go` - Snippet extraction utilities
- `utility/ffmpeg/ffmpeg.go` - Fixed WAV encoding in ChopOneSegment

---

## Test Results

### Test Case: Jude 1:1 (Second Half)
- **Original Audio**: 1984 recording
- **Target Text**: 2011 USX version
- **Segment**: "who are loved in God the Father and kept for Jesus Christ"
- **Boundaries**: 9.620-12.700s (high-confidence boundaries)

### Quality Assessment
- **TTS Generation**: ✅ Good
- **Boundary Detection**: ✅ Good
- **Audio Stitching**: ✅ Good
- **Voice Conversion**: ❌ Unacceptable (intelligible but robotic)
- **Prosody Matching**: ⏸️ Temporarily disabled

---

## Technical Decisions

1. **Go-Python Integration Pattern**: Using `stdio_exec` for subprocess communication (consistent with MMS modules)

2. **Conservative Voice Conversion**: Using pitch shift + loudness only (avoiding Griffin-Lim artifacts)

3. **Reference Segment Extraction**: Using exact segment being replaced as reference (better than whole chapter)

4. **Prosody Matching Disabled**: Focusing on voice conversion quality first

---

## Next Session Goals

1. **So-VITS-SVC Integration**:
   - Research installation and setup
   - Create training pipeline for speaker models
   - Integrate inference into voice conversion adapter

2. **Quality Validation**:
   - Train model on Jude speaker corpus
   - Test voice conversion quality
   - Iterate on training parameters if needed

3. **Re-enable Prosody Matching**:
   - Once voice conversion quality is acceptable
   - Fine-tune prosody matching parameters

---

## Notes for Future Reference

- MMS TTS does not support speaker embeddings natively
- DSP-based voice conversion insufficient for TTS→human conversion
- Segment boundary detection is a valuable building block
- So-VITS-SVC chosen over FCBH RVC for maturity and support
- Current voice conversion produces intelligible but unacceptable results

