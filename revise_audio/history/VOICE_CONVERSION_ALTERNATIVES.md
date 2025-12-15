# Voice Conversion Alternatives Analysis

**Status**: Current DSP-based approach (pitch shift + loudness) produces intelligible but unacceptable results.

**Problem**: Simple pitch shifting and loudness normalization cannot convert TTS-generated robotic audio into natural-sounding human speech that matches the original speaker.

---

## Option 1: Use Corpus Snippets Directly (RECOMMENDED - Short Term)

**Approach**: Instead of TTS + voice conversion, use actual audio snippets found in the corpus.

**Implementation**:
- Corpus search already finds matching words/phrases
- Extract snippets and use them directly (no conversion needed if same speaker)
- Only convert if different speaker (using one of the other options below)

**Pros**:
- ✅ **Zero quality loss** - uses actual human speech
- ✅ **No training required** - works immediately
- ✅ **Natural prosody** - already matches the context
- ✅ **Best match possible** - it's the actual speaker

**Cons**:
- ⚠️ Requires finding good matches in corpus (may not always be available)
- ⚠️ Different speaker snippets still need conversion

**Effort**: Minimal - corpus search already implemented

**Recommendation**: **Use this as primary approach, fall back to TTS only when corpus has no matches.**

---

## Option 2: Formant Shifting + Enhanced Spectral Matching

**Approach**: Add formant shifting (vocal tract characteristics) to current pitch + loudness approach.

**Technical Details**:
- Formants (F1, F2, F3) control vowel quality and voice timbre
- Use LPC (Linear Predictive Coding) to extract formants
- Shift formants to match target speaker's vocal tract characteristics
- Combine with improved spectral envelope matching (without Griffin-Lim reconstruction)

**Libraries**:
- `parselmouth` (Praat bindings) for formant extraction
- `pyworld` for LPC analysis
- `scipy` for signal processing

**Pros**:
- ✅ Still pure DSP - no training needed
- ✅ Captures more voice characteristics than pitch alone
- ✅ Can improve quality over current approach

**Cons**:
- ⚠️ May still not match full RVC quality
- ⚠️ More complex implementation
- ⚠️ Can introduce artifacts if not careful

**Effort**: Medium (1-2 days implementation)

**Example Code**:
```python
# Pseudo-code
def shift_formants(audio, target_formants, sample_rate):
    # Extract formants using LPC
    formants = extract_formants(audio, sample_rate)
    
    # Calculate shift ratios
    f1_ratio = target_formants[0] / formants[0]
    f2_ratio = target_formants[1] / formants[1]
    f3_ratio = target_formants[2] / formants[2]
    
    # Apply formant shifting using phase vocoder
    shifted = formant_shift(audio, [f1_ratio, f2_ratio, f3_ratio])
    
    return shifted
```

---

## Option 3: So-VITS-SVC (Open Source RVC)

**Approach**: Use So-VITS-SVC, a mature open-source RVC implementation.

**Why So-VITS-SVC over FCBH RVC**:
- More mature community support
- Better documentation
- Pre-trained base models available
- Easier to use for inference (training still needed per speaker)

**Technical Details**:
- Similar to FCBH RVC: HuBERT content encoder + F0 + speaker embedding
- More polished inference pipeline
- Better vocoder integration

**Pros**:
- ✅ Better supported than FCBH RVC
- ✅ Similar architecture (HuBERT-based)
- ✅ Better documentation/examples
- ✅ Can train per-speaker models

**Cons**:
- ⚠️ Still requires training per speaker (but corpus is available)
- ⚠️ Different codebase to integrate
- ⚠️ Training time (but one-time per speaker)

**Effort**: High (3-5 days for integration + training)

**Repository**: https://github.com/svc-develop-team/so-vits-svc

---

## Option 4: Train FCBH RVC Model for This Speaker

**Approach**: Use existing FCBH RVC infrastructure to train a model for the target speaker.

**Requirements**:
- Audio corpus for the speaker (available from database)
- Training time (several hours to days depending on data)
- Fix speechbrain/torchaudio compatibility issues first

**Pros**:
- ✅ Uses existing codebase
- ✅ Designed for minority languages
- ✅ Once trained, can be reused

**Cons**:
- ⚠️ Requires fixing compatibility issues first
- ⚠️ Training time investment
- ⚠️ Need to validate training quality

**Effort**: High (2-3 days to fix issues + training time)

**Implementation Steps**:
1. Fix speechbrain/torchaudio compatibility in conda env
2. Extract training data for speaker from corpus
3. Train RVC model using FCBH infrastructure
4. Integrate trained model into voice conversion pipeline

---

## Option 5: OpenVoice / XTTS-v2 (Zero-Shot)

**Approach**: Use zero-shot voice cloning systems that don't require training.

**OpenVoice**:
- Zero-shot voice cloning
- Better prosody control
- Multilingual support

**XTTS-v2**:
- Coqui TTS's zero-shot system
- Limited to 17 languages (English is supported)

**Pros**:
- ✅ No training required
- ✅ Works immediately with reference audio
- ✅ Better than DSP approaches

**Cons**:
- ⚠️ May not be optimized for minority languages
- ⚠️ Different integration pattern
- ⚠️ Quality may not match trained RVC

**Effort**: Medium-High (2-3 days for integration)

**Example** (OpenVoice):
```python
from openvoice import se_extractor, base_speaker, ToneColorConverter

# Extract speaker embedding
tone_color_converter = ToneColorConverter(...)
target_se, audio_name = se_extractor.get_se(source_audio, ...)

# Convert
src_path = "tts_output.wav"
output_path = "converted.wav"
tone_color_converter.convert(
    audio_src_path=src_path,
    src_se=source_se,
    tgt_se=target_se,
    output_path=output_path,
)
```

---

## Option 6: Neural Vocoder + Spectral Envelope Transfer

**Approach**: Use neural vocoder (HiFi-GAN/BigVGAN) instead of Griffin-Lim for reconstruction.

**Current Problem**: Griffin-Lim reconstruction introduces artifacts when doing spectral envelope transfer.

**Solution**: 
- Extract mel spectrogram from source
- Modify spectral envelope
- Use neural vocoder to reconstruct (better than Griffin-Lim)

**Vocoders Available**:
- HiFi-GAN (from FCBH codebase)
- BigVGAN (better quality, more recent)

**Pros**:
- ✅ Better reconstruction quality than Griffin-Lim
- ✅ Can reuse existing vocoder infrastructure
- ✅ Less training needed (vocoders are general-purpose)

**Cons**:
- ⚠️ Still requires spectral envelope matching logic
- ⚠️ May need vocoder model loading

**Effort**: Medium (1-2 days)

---

## Recommendation: Hybrid Approach

**Phase 1 (Immediate)**: 
1. **Use corpus snippets directly** when available (Option 1)
2. **Add formant shifting** to current DSP approach for better quality (Option 2)

**Phase 2 (Short-term)**:
3. **Train FCBH RVC model** for target speakers (Option 4)
   - Fix compatibility issues
   - Train on available corpus
   - Use trained model for conversion

**Phase 3 (If needed)**:
4. **Evaluate So-VITS-SVC** if FCBH RVC doesn't meet quality requirements (Option 3)
5. **Consider OpenVoice** for zero-shot scenarios (Option 5)

---

## Quick Win: Formant Shifting Implementation

Should I implement Option 2 (Formant Shifting) next? It's the quickest path to improved quality while keeping the DSP approach.

