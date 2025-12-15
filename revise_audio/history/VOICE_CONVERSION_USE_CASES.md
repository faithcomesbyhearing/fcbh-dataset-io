# Voice Conversion Use Cases

## Overview

The audio revision system needs to handle two distinct scenarios that require different approaches:

---

## Use Case 1: Text-Driven Revisions (Current Focus)

**Scenario**: Text is available, need to generate new audio content.

**Example**: 
- Original recording: "Jude, a servant of Jesus Christ and a brother of James"
- Revised text: "Jude, a servant of Jesus Christ and a brother of James" (with corrected words)
- Need to generate corrected audio

**Workflow**:
1. Generate TTS audio from revised text (MMS TTS)
2. Voice convert TTS output to match original speaker (So-VITS-SVC)
3. Match prosody to surrounding context
4. Replace segment in original audio

**Requirements**:
- TTS system (MMS TTS) ✅
- Voice conversion (So-VITS-SVC) ⏳
- Prosody matching ✅
- Audio stitching ✅

**Status**: In progress - So-VITS-SVC integration needed

---

## Use Case 2: Audio-Only Revisions (No Text Available)

**Scenario**: No writing system available, have recorded audio snippets from native speakers.

**Example**:
- Original recording: [speaker A saying phrase X]
- Native speaker recorded: [speaker B saying corrected phrase Y]
- Need to replace phrase X with phrase Y, but in speaker A's voice

**Workflow**:
1. Record audio snippet from native speaker (no text needed)
2. Voice convert snippet to match original speaker (So-VITS-SVC)
3. Match prosody to surrounding context
4. Replace segment in original audio

**Requirements**:
- Voice conversion (So-VITS-SVC) ⏳
- Prosody matching ✅
- Audio stitching ✅
- **No TTS needed** ✅

**Status**: Same voice conversion system (So-VITS-SVC) will work for this

---

## Key Insight: So-VITS-SVC Works for Both

**So-VITS-SVC is a general voice conversion system** - it doesn't care about the source of the input audio:

- **Input**: Any audio (TTS-generated OR human-recorded)
- **Target**: Speaker embedding from original recording
- **Output**: Audio in target speaker's voice

### Why Audio-Only Might Actually Work Better

**Advantages of audio snippets over TTS**:
1. ✅ **Already natural speech** - no robotic artifacts to fix
2. ✅ **Natural prosody** - native speaker's natural rhythm and intonation
3. ✅ **Better quality input** - human speech is easier to convert than TTS
4. ✅ **Language-agnostic** - works for any language, even without writing system

**The voice conversion task is simpler**:
- TTS → Human: Need to fix robotic artifacts + change voice
- Human → Human: Only need to change voice (easier!)

---

## Implementation Considerations

### For Use Case 1 (Text-Driven)

```python
# Current approach
tts_audio = mms_tts.generate(text)
converted_audio = so_vits_svc.convert(tts_audio, target_speaker_embedding)
prosody_matched = prosody_match(converted_audio, reference_context)
final_audio = stitch(original_audio, prosody_matched, segment)
```

### For Use Case 2 (Audio-Only)

```python
# Simpler approach
recorded_snippet = load_audio("native_speaker_recording.wav")
converted_audio = so_vits_svc.convert(recorded_snippet, target_speaker_embedding)
prosody_matched = prosody_match(converted_audio, reference_context)
final_audio = stitch(original_audio, prosody_matched, segment)
```

**Note**: Same So-VITS-SVC conversion step, just different input source!

---

## Workflow Comparison

### Text-Driven (Use Case 1)
```
Revised Text
    ↓
MMS TTS
    ↓
[TTS Audio - robotic]
    ↓
So-VITS-SVC (fixes robotic + changes voice)
    ↓
[Converted Audio - natural, correct voice]
    ↓
Prosody Matching
    ↓
Audio Stitching
    ↓
Final Revised Audio
```

### Audio-Only (Use Case 2)
```
Native Speaker Recording
    ↓
[Recorded Audio - natural, wrong voice]
    ↓
So-VITS-SVC (changes voice only)
    ↓
[Converted Audio - natural, correct voice]
    ↓
Prosody Matching
    ↓
Audio Stitching
    ↓
Final Revised Audio
```

**Key Difference**: Audio-only skips TTS step and starts with natural speech, making voice conversion easier.

---

## Recommendations

### For Text-Driven Revisions (Use Case 1)
- **So-VITS-SVC is still the right choice**
- May need higher quality TTS or post-processing
- Prosody matching is critical (TTS prosody often needs adjustment)

### For Audio-Only Revisions (Use Case 2)
- **So-VITS-SVC is perfect for this**
- This is the "standard" voice conversion use case
- Should work better than text-driven (easier conversion task)
- Prosody matching still valuable but less critical (input already has natural prosody)

---

## Implementation Status

**Current Implementation**:
- ✅ Voice conversion infrastructure (Go-Python interface)
- ✅ Prosody matching
- ✅ Audio stitching
- ⏳ So-VITS-SVC integration (needed for both use cases)

**After So-VITS-SVC Integration**:
- Both use cases will be supported
- Same voice conversion code path
- Different input sources (TTS vs recorded audio)

---

## Future Considerations

### Hybrid Approach
For languages with partial writing systems:
- Use TTS for words that can be written
- Use recorded snippets for words that can't be written
- Both go through same voice conversion pipeline

### Quality Comparison
Once So-VITS-SVC is integrated, we should:
- Compare quality: TTS+VC vs Recorded+VC
- Document which works better for different scenarios
- Optimize workflow based on results

---

## Summary

**So-VITS-SVC is the right choice for BOTH use cases**:
- ✅ Text-driven: TTS → VC → Prosody → Stitch
- ✅ Audio-only: Recorded → VC → Prosody → Stitch

**Audio-only use case is actually simpler** because:
- Input is already natural speech (no robotic artifacts)
- Voice conversion task is easier (human→human vs TTS→human)
- Should produce better quality results

The same So-VITS-SVC integration will support both workflows!

