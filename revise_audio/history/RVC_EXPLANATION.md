# RVC and Voice Conversion Explained

## What is RVC (Retrieval-based Voice Conversion)?

**RVC** stands for **Retrieval-based Voice Conversion**. It's a neural voice cloning technique that converts one speaker's voice to sound like another speaker while preserving the linguistic content.

### How RVC Works (High-Level)

1. **Content Extraction**: Uses a pre-trained model (like HuBERT or WavLM) to extract linguistic/phonetic content from source audio, stripping away speaker identity
2. **Speaker Embedding**: Extracts a compact representation (embedding) of the target speaker's voice characteristics
3. **F0 (Pitch) Extraction**: Separates pitch information from the source
4. **Conversion**: A neural network combines:
   - Content features (what is said)
   - Target speaker embedding (who says it)
   - Source F0 (how it's said - pitch contour)
5. **Vocoder**: Converts the modified features back to audio waveform

### Why "Retrieval-based"?

The name comes from using a **FAISS index** to retrieve similar voice characteristics from a training corpus, helping the model match the target speaker's voice more accurately.

---

## Formant Shifting Explained

**Formants** are resonant frequencies of the vocal tract that determine vowel quality and voice timbre. They're what makes an "ah" sound different from an "ee" sound, and what makes one person's voice sound different from another's.

### What Formants Are

- **F1 (First Formant)**: Related to tongue height (high/low vowels)
- **F2 (Second Formant)**: Related to tongue front/back position
- **F3 (Third Formant)**: Related to lip rounding and overall timbre

### Formant Shifting

Formant shifting adjusts these frequencies to match a target speaker's vocal tract characteristics. It's more sophisticated than pitch shifting because it changes the **timbre** (tone color) of the voice, not just the pitch.

**Example**:
- Pitch shifting: Makes a voice higher/lower (like changing octave on a piano)
- Formant shifting: Makes a voice sound like a different person (like changing the instrument)

**Limitation**: Formant shifting alone still can't fully convert TTS to natural human speech - it's a DSP technique, not neural modeling.

---

## Option 3: FCBH RVC vs Option 4: So-VITS-SVC

### Similarities

Both are **RVC implementations** with very similar architectures:
- Both use HuBERT/WavLM for content extraction
- Both use F0 extraction (RMVPE, pyworld, etc.)
- Both use speaker embeddings
- Both train neural networks to convert voice
- Both use vocoders for audio reconstruction

### Key Differences

#### Option 3: FCBH RVC (Custom Implementation)

**What it is**: A custom RVC implementation built specifically for the FCBH codebase.

**Architecture**:
- Custom `FCBHRVCModel` (PyTorch module)
- Integrated with FCBH's existing infrastructure
- Uses FCBH's feature extraction pipeline
- Designed for minority languages

**Pros**:
- ✅ Already in your codebase (no new dependencies)
- ✅ Integrated with existing tools
- ✅ Designed for minority languages
- ✅ You have full control over the code

**Cons**:
- ⚠️ Less mature than So-VITS-SVC
- ⚠️ Smaller community (harder to find help/examples)
- ⚠️ May have bugs or limitations
- ⚠️ Currently has compatibility issues (speechbrain/torchaudio)

**Status**: Code exists but needs:
- Fix compatibility issues
- Training on your speaker data
- Validation that it works well

---

#### Option 4: So-VITS-SVC (Open Source Community Project)

**What it is**: A mature, open-source RVC implementation with large community support.

**Architecture**:
- Well-tested RVC implementation
- Active GitHub repository with many contributors
- Extensive documentation and examples
- Pre-trained base models available

**Pros**:
- ✅ **Mature and battle-tested** (used by many people)
- ✅ **Better documentation** and examples
- ✅ **Active community** (easier to get help)
- ✅ **Pre-trained models** available (can fine-tune faster)
- ✅ **Regular updates** and bug fixes
- ✅ **Better inference pipeline** (more polished)

**Cons**:
- ⚠️ **Different codebase** (need to integrate separately)
- ⚠️ **May not be optimized for minority languages** (though HuBERT is cross-lingual)
- ⚠️ **Less control** over internals
- ⚠️ **Still requires training** per speaker

**Status**: External project, would need to:
- Install as separate dependency
- Integrate into your workflow
- Train on your speaker data

---

## Which Should You Choose?

### Recommendation: **Option 4 (So-VITS-SVC)** for these reasons:

1. **Maturity**: So-VITS-SVC is more mature and has been tested by many users
2. **Community Support**: Easier to find solutions to problems
3. **Documentation**: Better docs make integration faster
4. **Quality**: More polished inference pipeline likely means better results
5. **Time to Working**: Despite being external, likely faster to get working than fixing FCBH RVC

### However, consider Option 3 (FCBH RVC) if:

- You want to stay within your existing codebase
- You have specific requirements for minority languages that FCBH RVC addresses
- You want full control over the implementation
- You're willing to invest time fixing compatibility issues

---

## Recent AI Improvements in Voice Cloning

You're correct that recent AI has seen significant improvements:

1. **Better Content Encoders**: HuBERT, WavLM are much better than older models
2. **Better F0 Extraction**: RMVPE is state-of-the-art
3. **Better Vocoders**: HiFi-GAN, BigVGAN produce higher quality audio
4. **Better Training**: More stable training procedures

**Both FCBH RVC and So-VITS-SVC use these modern techniques**, so either should give you good results. The question is more about:
- Which is easier to get working?
- Which has better support?
- Which fits your workflow better?

---

## Next Steps Recommendation

Given your constraints:
1. **Corpus snippets**: Tried, word boundaries too imprecise
2. **TTS needed**: For generating new content
3. **Modeling needed**: DSP approaches insufficient
4. **Minority languages**: XTTS-v2 not suitable

**I recommend: So-VITS-SVC (Option 4)**

**Why**:
- Most likely to work well out of the box
- Faster to integrate than fixing FCBH RVC
- Better community support for troubleshooting
- Uses same modern techniques (HuBERT, RMVPE) so quality should be similar

**Implementation Plan**:
1. Install So-VITS-SVC as a Python package/service
2. Train a model on your speaker's corpus data
3. Integrate inference into your `revise_audio` pipeline
4. Use it to convert TTS output to match original speaker

Would you like me to:
1. Research So-VITS-SVC integration approach?
2. Start implementing So-VITS-SVC integration?
3. Or try to fix FCBH RVC compatibility issues first?
