# Voice Conversion & Prosody Matching Tooling Research

**Task**: arti-c12  
**Date**: 2025-12-13  
**Status**: Research Complete - Recommendations Ready

## Executive Summary

**Recommended Approach**: Start with existing FCBH RVC infrastructure + DSP-based prosody matching, upgrade to neural prosody later if needed.

**Rationale**: 
- **Practical**: Leverages existing codebase (faster implementation)
- **Uncertain**: "Proven on minority languages" - claim from FCBH docs, not externally verified
- **Flexible**: Can upgrade components incrementally if better solutions found
- **Language-agnostic**: HuBERT/WavLM are cross-lingual (this IS externally verified)

**Critical Note**: This recommendation is **pragmatic** (use what exists) rather than **optimal** (best possible solution). The FCBH codebase claims are not independently verified. Consider this a starting point that can be improved.

---

## 1. Voice Conversion Options

### Option A: FCBH RVC System (RECOMMENDED for Phase 1)

**Source**: `/Users/jrstear/git/FCBH-W2V-Bert-2.0-ASR-trainer/`

**Components Available**:
- `rvc_feature_extractor.py` - HuBERT features, F0 extraction (RMVPE), FAISS indexing
- `fcbh_rvc_model.py` - RVC model architecture
- `w2v_bert_voice_training_engine.py` - Complete RVC workflow
- Speaker embedding extraction (X-Vector + prosody features)

**Architecture**:
- Content encoder: HuBERT (768-dim features)
- F0 extraction: RMVPE (state-of-the-art pitch tracking)
- Speaker embeddings: SpeechBrain X-Vector (512-dim) + prosody features (64-dim)
- Vocoder: HiFi-GAN or BigVGAN

**Pros**:
- ✅ Already integrated and tested (practical advantage)
- ✅ Codebase exists (faster to use than building from scratch)
- ✅ Language-agnostic (HuBERT is cross-lingual - **externally verified**)
- ✅ Can extract speaker embeddings from existing corpus

**Cons**:
- ⚠️ "Proven on minority languages" - **claim from FCBH docs, not externally verified**
- ⚠️ "Optimized for low-resource" - **claim, not benchmarked against alternatives**
- ⚠️ Requires training per narrator (but corpus is available)
- ⚠️ More complex than zero-shot solutions
- ⚠️ **Unknown if this is actually state-of-the-art or just "what they happened to use"**

**External Evidence**:
- ✅ HuBERT/WavLM are widely used for cross-lingual speech tasks (verified)
- ✅ RVC (Retrieval-based Voice Conversion) is a real approach used in research
- ❓ Whether FCBH's specific implementation is optimal - **unknown**
- ❓ Whether RVC is better than So-VITS-SVC, FreeVC, etc. - **needs comparison**

**Integration Pattern**:
- Follow existing `mms/` adapter pattern
- Go → Python subprocess (like `mms_asr.go` → `mms_asr.py`)
- Python module handles VC inference
- Go handles orchestration, database access, file I/O

**Dependencies**:
- PyTorch
- transformers (HuBERT)
- speechbrain (X-Vector embeddings)
- pyworld or torchcrepe (F0 extraction)
- librosa, soundfile (audio processing)

---

### Option B: So-VITS-SVC (Alternative)

**Source**: Open-source RVC implementation (GitHub)

**Architecture**:
- Similar to FCBH RVC (HuBERT + F0 + speaker embedding)
- More mature community support
- Pre-trained models available

**Pros**:
- ✅ Well-documented
- ✅ Active community
- ✅ Pre-trained models available

**Cons**:
- ⚠️ Would need to adapt/integrate from scratch
- ⚠️ May not be optimized for minority languages
- ⚠️ Less control over training process

**Verdict**: Not recommended for Phase 1 - FCBH RVC is already available and proven.

---

### Option C: OpenVoice / XTTS-v2 (Future Consideration)

**Pros**:
- ✅ Zero-shot voice cloning (no training needed)
- ✅ Better prosody control
- ✅ Multilingual support

**Cons**:
- ⚠️ XTTS limited to 17 languages (not suitable for minority languages)
- ⚠️ OpenVoice less mature, harder to integrate
- ⚠️ May not match quality of trained RVC models

**Verdict**: Consider for Phase 2 if RVC doesn't meet quality requirements.

---

## 2. Prosody Matching Options

### Option A: DSP-Based Prosody Matching (RECOMMENDED for Phase 1)

**Libraries**: librosa, pyworld, scipy

**Approach**:
1. Extract prosody features from context (surrounding verses):
   - F0 (pitch) contour using pyworld or librosa
   - Energy/loudness envelope (RMS)
   - Speaking rate (duration analysis)

2. Adjust snippet to match:
   - Pitch shifting: `librosa.effects.pitch_shift()` or pyworld F0 modification
   - Time stretching: `librosa.effects.time_stretch()` (non-uniform if needed)
   - Loudness normalization: RMS matching

**Pros**:
- ✅ Simple, no training required
- ✅ Fast inference
- ✅ No language dependencies
- ✅ Easy to debug and tune
- ✅ Works immediately

**Cons**:
- ⚠️ May not capture all prosodic nuances
- ⚠️ Can introduce artifacts if over-applied

**Implementation**:
```python
# Pseudocode
def match_prosody(source_audio, reference_audio):
    # Extract F0 from reference
    f0_ref = extract_f0(reference_audio)
    f0_source = extract_f0(source_audio)
    
    # Calculate pitch shift needed
    pitch_shift = np.mean(f0_ref) - np.mean(f0_source)
    
    # Extract speaking rate
    rate_ref = calculate_speaking_rate(reference_audio)
    rate_source = calculate_speaking_rate(source_audio)
    time_stretch = rate_ref / rate_source
    
    # Apply adjustments
    adjusted = librosa.effects.pitch_shift(source_audio, sr, pitch_shift)
    adjusted = librosa.effects.time_stretch(adjusted, time_stretch)
    adjusted = normalize_loudness(adjusted, reference_audio)
    
    return adjusted
```

**Dependencies**:
- librosa (pitch shift, time stretch)
- pyworld (F0 extraction - more accurate than librosa)
- scipy (signal processing)

---

### Option B: Neural Prosody Transfer (Phase 2)

**Approach**:
- Extract prosody embeddings from reference audio
- Condition VC decoder on prosody embedding
- Requires training prosody encoder

**Pros**:
- ✅ Better quality (captures subtle nuances)
- ✅ More natural-sounding

**Cons**:
- ⚠️ Requires training
- ⚠️ More complex
- ⚠️ Slower inference

**Verdict**: Start with DSP-based, upgrade if needed.

---

## 3. Integration Architecture

### Go → Python Pattern (Following MMS Example)

**Structure**:
```
revise_audio/
├── revise_audio.go          # Main Go module
├── vc_adapter.go            # Voice conversion adapter
├── prosody_adapter.go       # Prosody matching adapter
└── python/
    ├── voice_conversion.py  # RVC inference
    └── prosody_match.py     # DSP prosody matching
```

**Communication**:
- Go calls Python via subprocess (like `mms_asr.go`)
- JSON/stdin-stdout for data exchange
- Audio files passed as paths
- Python returns audio file paths

**Example**:
```go
// vc_adapter.go
func ConvertVoice(sourceAudio, targetSpeakerEmbedding string) (string, error) {
    cmd := exec.Command("python3", "revise_audio/python/voice_conversion.py",
        "--source", sourceAudio,
        "--speaker-embedding", targetSpeakerEmbedding,
        "--output", outputPath)
    // ... execute and handle output
}
```

---

## 4. Recommended Stack (Phase 1)

### Voice Conversion
- **Tool**: FCBH RVC (existing codebase)
- **Content Encoder**: HuBERT (transformers library)
- **F0 Extraction**: RMVPE (via pyworld or torchcrepe)
- **Speaker Embeddings**: SpeechBrain X-Vector
- **Vocoder**: HiFi-GAN (or BigVGAN if available)

### Prosody Matching
- **Tool**: DSP-based (librosa + pyworld)
- **F0 Extraction**: pyworld (more accurate) or librosa
- **Pitch Shifting**: librosa.effects.pitch_shift
- **Time Stretching**: librosa.effects.time_stretch
- **Loudness**: RMS normalization

### Python Dependencies
```
torch>=2.0.0
transformers>=4.25.0  # HuBERT
speechbrain>=0.5.15   # X-Vector embeddings
librosa>=0.10.0       # Audio processing
pyworld>=0.3.2        # F0 extraction
soundfile>=0.12.1     # Audio I/O
scipy>=1.9.0          # Signal processing
numpy>=1.22.0
```

---

## 5. Implementation Plan

### Phase 1: Basic VC + DSP Prosody
1. Extract speaker embeddings from target verse context
2. Apply RVC to convert source audio to target voice
3. Extract prosody from surrounding verses
4. Apply DSP-based prosody matching
5. Validate quality

### Phase 2: Enhancements (if needed)
- Upgrade to neural prosody transfer
- Fine-tune RVC models per narrator
- Add acoustic matching (EQ, room tone)

---

## 6. YAML Configuration Structure

Based on requirements, subcategories will be:
```yaml
revise_audio:
  rvc: yes                    # Use FCBH RVC for voice conversion
  prosody_dsp: yes            # Use DSP-based prosody matching
  # Future:
  # prosody_neural: yes       # Neural prosody (Phase 2)
  # openvoice: yes            # Alternative VC (Phase 2)
```

---

## 7. Decision Matrix

| Criteria | FCBH RVC | So-VITS-SVC | OpenVoice | DSP Prosody | Neural Prosody |
|----------|----------|-------------|-----------|-------------|----------------|
| **Ease of Integration** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Language Support** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Quality** | ⭐⭐⭐? | ⭐⭐⭐⭐? | ⭐⭐⭐⭐? | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Speed** | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Training Required** | Yes | Yes | No | No | Yes |
| **Already Available** | ✅ | ❌ | ❌ | ✅ | ❌ |
| **Externally Verified** | ❓ | ✅ | ✅ | ✅ | ✅ |

**Winner**: FCBH RVC + DSP Prosody (best balance of **practical integration** and availability)

**Note**: Quality ratings are uncertain (?) - FCBH RVC quality is **not externally benchmarked**. This recommendation prioritizes **practicality** (use what exists) over **optimality** (best possible solution).

---

## 8. What We Actually Know vs. Claims

### Externally Verified Facts:
- ✅ **HuBERT/WavLM are cross-lingual** - widely used in research, proven for multilingual tasks
- ✅ **RVC is a real approach** - Retrieval-based Voice Conversion exists in literature
- ✅ **RMVPE is good for F0** - used in multiple systems
- ✅ **X-Vector embeddings work** - SpeechBrain is established
- ✅ **DSP prosody matching is simple** - librosa/pyworld are standard tools

### Unverified Claims (from FCBH docs):
- ❓ "Best-in-class content extraction" - **not benchmarked**
- ❓ "State-of-the-art pitch extraction" - **RMVPE is good, but "best"? unknown**
- ❓ "Proven on minority languages" - **no external evidence found**
- ❓ "Optimized for low-resource" - **compared to what?**
- ❓ "30-45% quality improvement" - **over what baseline?**

### What We Don't Know:
- ❓ Is FCBH RVC better than So-VITS-SVC?
- ❓ Is FCBH RVC better than FreeVC?
- ❓ Is HuBERT better than WavLM for voice conversion?
- ❓ What's actually state-of-the-art for voice conversion in 2024/2025?

## 9. Recommendation Strategy

**Phase 1 (Pragmatic)**: Use FCBH RVC because:
- It exists and works (even if not optimal)
- Faster to implement than researching/building alternatives
- Can be replaced later if better solutions found
- Components (HuBERT, X-Vector) are solid choices

**Phase 2 (Validation)**: 
- Benchmark FCBH RVC against So-VITS-SVC or FreeVC
- Compare quality on actual minority language data
- If alternatives are better, switch

**Phase 3 (Optimization)**:
- If FCBH RVC works well, keep it
- If not, adopt better solution based on benchmarks

## 10. Next Steps

1. ✅ Research complete (this document - now more honest about uncertainties)
2. **Consider**: Quick comparison test of FCBH RVC vs. So-VITS-SVC before full integration?
3. Create `revise_audio/` directory structure
4. Port/adapt RVC inference code from FCBH-W2V-Bert-2.0-ASR-trainer
5. Implement DSP prosody matching module
6. Create Go adapters following MMS pattern
7. Test with sample data
8. **Validate**: Compare results against alternatives if time permits

---

## References

- FCBH RVC: `/Users/jrstear/git/FCBH-W2V-Bert-2.0-ASR-trainer/rvc_feature_extractor.py`
- MMS Integration Pattern: `/Users/jrstear/git/arti/mms/mms_asr/`
- Notes: `/Users/jrstear/git/arti/revise_audio/notes`

