# RVC (Retrieval-based Voice Conversion) - Detailed Explanation

**Date**: 2024-12-15  
**Purpose**: Comprehensive explanation of what RVC is, how it works, and how it differs from TTS

---

## What is RVC?

**RVC** = **Retrieval-based Voice Conversion**

RVC is a **voice cloning** technique that converts one speaker's voice to sound like another speaker while preserving:
- ✅ **Linguistic content** (what is said)
- ✅ **Prosody** (how it's said - pitch, rhythm, timing)
- ✅ **Emotional tone** (if present in source)

**Key Point**: RVC is **NOT TTS** (Text-to-Speech). It's **voice conversion** - it takes existing audio and changes the voice.

---

## RVC vs TTS vs Voice Cloning

### TTS (Text-to-Speech)
- **Input**: Text
- **Output**: Speech audio
- **Purpose**: Generate speech from text
- **Example**: "Hello world" → Audio saying "Hello world"

### Voice Cloning (General Term)
- **Input**: Audio from one speaker
- **Output**: Same audio, different speaker's voice
- **Purpose**: Make audio sound like a different person said it
- **Example**: Your voice saying "Hello" → Sounds like celebrity saying "Hello"

### RVC (Retrieval-based Voice Conversion)
- **Type**: Voice cloning technique
- **Method**: Uses retrieval mechanism to find similar voice characteristics
- **Input**: Audio from source speaker
- **Output**: Audio in target speaker's voice
- **Example**: TTS output → Converted to sound like original narrator

---

## How RVC Works (Step-by-Step)

### Architecture Overview

```
Source Audio
    ↓
[1. Content Extraction] → Linguistic features (HuBERT/WavLM)
    ↓
[2. F0 Extraction] → Pitch contour (RMVPE/pyworld)
    ↓
[3. Speaker Embedding] → Target speaker characteristics (X-Vector)
    ↓
[4. Retrieval] → Find similar voice segments (FAISS index)
    ↓
[5. Neural Conversion] → Combine features + speaker embedding
    ↓
[6. Vocoder] → Generate audio waveform
    ↓
Converted Audio (target speaker's voice)
```

### Detailed Component Explanation

#### 1. Content Extraction (What is Said)

**Purpose**: Extract linguistic/phonetic content, removing speaker identity

**How**:
- Uses pre-trained models like **HuBERT** or **WavLM**
- These are self-supervised models trained on massive audio datasets
- Extract 768-dimensional feature vectors representing speech content
- **Language-agnostic**: Works across languages (cross-lingual)

**Output**: Content features (what words/phonemes are present)

**Example**:
- Input: "Hello" in Speaker A's voice
- Output: Features representing "Hello" (without Speaker A's voice characteristics)

#### 2. F0 Extraction (Pitch - How it's Said)

**Purpose**: Extract fundamental frequency (pitch) contour

**How**:
- Uses **RMVPE** (Robust Model for Voice Pitch Extraction) - state-of-the-art
- Or alternatives: pyworld, torchcrepe, Harvest, DIO
- Extracts pitch at each time frame
- Preserves prosody (intonation, rhythm)

**Output**: F0 contour (pitch over time)

**Example**:
- Input: Rising intonation on "Hello?"
- Output: F0 values showing the pitch rise

#### 3. Speaker Embedding (Who Says It)

**Purpose**: Capture target speaker's voice characteristics

**How**:
- Uses **X-Vector** embeddings (SpeechBrain ECAPA-TDNN)
- Extracts 512-dimensional vector representing:
  - Vocal tract characteristics
  - Timbre (tone color)
  - Voice quality
  - Speaking style
- **FCBH Enhancement**: Adds 64D prosody features (30-45% quality improvement)

**Output**: Speaker embedding vector

**Example**:
- Input: 30 minutes of target speaker's audio
- Output: 512D vector representing that speaker's voice

#### 4. Retrieval Mechanism (The "Retrieval" in RVC)

**Purpose**: Find similar voice segments from training corpus

**How**:
- Uses **FAISS** (Facebook AI Similarity Search) index
- Pre-computes features from training corpus
- During inference, searches for similar content features
- Retrieves matching voice characteristics
- Helps model match target speaker more accurately

**Why "Retrieval-based"?**
- Instead of purely neural conversion, it retrieves actual examples
- More accurate voice matching
- Better naturalness

**Output**: Retrieved voice segments/features

#### 5. Neural Conversion (The Magic)

**Purpose**: Combine all features to create converted voice

**How**:
- Neural network (encoder-decoder architecture)
- Takes:
  - Content features (what to say)
  - Speaker embedding (who says it)
  - F0 contour (how to say it - pitch)
  - Retrieved features (voice matching)
- Learns to combine these into target speaker's voice
- Outputs modified mel-spectrogram

**Training**:
- Trained on pairs: (source audio, target speaker audio)
- Learns mapping: source features → target voice
- Requires 30+ minutes of target speaker audio (ideally)

**Output**: Modified mel-spectrogram in target voice

#### 6. Vocoder (Audio Generation)

**Purpose**: Convert mel-spectrogram back to audio waveform

**How**:
- Uses neural vocoder (HiFi-GAN, BigVGAN, or NSF-HiFiGAN)
- Converts frequency-domain features to time-domain audio
- Generates final audio waveform

**Output**: Audio file in target speaker's voice

---

## FCBH RVC vs So-VITS-SVC

### Similarities

Both are RVC implementations with nearly identical architectures:
- ✅ Content extraction (HuBERT/WavLM)
- ✅ F0 extraction (RMVPE)
- ✅ Speaker embeddings (X-Vector)
- ✅ FAISS retrieval
- ✅ Neural conversion network
- ✅ Vocoder synthesis

### Differences

| Aspect | FCBH RVC | So-VITS-SVC |
|--------|----------|-------------|
| **Maturity** | Custom, less tested | Mature, widely used |
| **Community** | Small (FCBH internal) | Large open-source |
| **Documentation** | Limited | Extensive |
| **Optimization** | Claims "low-resource" | General purpose |
| **Integration** | Already in codebase | External dependency |
| **Control** | Full control | Less control |
| **Support** | Internal only | Community support |

### FCBH RVC Specific Features

1. **Prosody-Enhanced Embeddings**: Adds 64D prosody features to X-Vector
   - 30-45% quality improvement claimed
   - F0 patterns, energy dynamics, speaking rate, voice quality

2. **Low-Resource Optimization**: Claims to work with less data
   - Functional: 300-500 samples
   - Good: 1000-2000 samples
   - Excellent: 3000-5000 samples

3. **ASR Integration**: Seamless integration with W2V-BERT ASR pipeline
   - Shared vocabulary
   - Unified folder structure
   - Auto-phoneme generation

---

## Is RVC Voice Cloning or TTS?

### RVC is Voice Cloning (NOT TTS)

**RVC**:
- ✅ Takes **audio input** (not text)
- ✅ Converts **existing speech** to different voice
- ✅ Preserves **prosody and timing**
- ✅ Can work with **any audio** (TTS output, recordings, etc.)

**TTS**:
- ✅ Takes **text input**
- ✅ Generates **new speech** from scratch
- ✅ May or may not match specific voice
- ✅ Requires text-to-speech model

### How They Work Together

**Typical Workflow**:
```
Text → [TTS] → Audio (generic voice)
              ↓
         [RVC] → Audio (target speaker's voice)
```

**For Your Use Case**:
1. Generate TTS from revised text (MMS TTS)
2. Convert TTS output to match original narrator (RVC)
3. Match prosody to surrounding context (Prosody Matching)
4. Stitch into original audio

---

## Training Requirements

### What RVC Needs

**Training Data**:
- **Minimum**: 30+ minutes of clean, diverse speech
- **Recommended**: 1-2 hours
- **Format**: WAV files, 16kHz, mono
- **Quality**: Clean, consistent speaker, minimal noise

**Training Process**:
1. Extract features (HuBERT, F0) from training data
2. Build FAISS index for retrieval
3. Train neural network on (source, target) pairs
4. Generate speaker embeddings
5. Save model checkpoint

**Training Time**:
- **CPU**: Very slow (days/weeks) - not recommended
- **GPU**: 2-8 hours depending on data size
- **Epochs**: 150-600 epochs typically

### Why Your 26 Segments Failed

**Your Data**: 26 verse segments (~4 minutes)
**Required**: 30+ minutes minimum

**Problem**:
- Insufficient data → model overfits
- Can't generalize to new phrases
- Produces artifacts
- Quality degradation

**Solution**:
- Collect more training data (30+ minutes)
- Or use corpus snippets when available (no training needed)

---

## RVC Use Cases

### 1. Voice Cloning (Your Use Case)
- Convert TTS output to match original narrator
- Preserve voice characteristics across revisions

### 2. Dubbing
- Convert speech to different speaker's voice
- Maintain original prosody and timing

### 3. Voice Restoration
- Restore damaged/old recordings
- Convert to modern voice while preserving content

### 4. Character Voice Creation
- Create consistent character voices
- Convert any speech to character's voice

---

## Advantages of RVC

### ✅ Pros

1. **Preserves Prosody**: Maintains natural rhythm, intonation, timing
2. **Language-Agnostic**: Works across languages (HuBERT is cross-lingual)
3. **Data Efficient**: Can work with 30+ minutes (vs. hours for some systems)
4. **High Quality**: Neural approach produces natural-sounding results
5. **Flexible Input**: Works with any audio (TTS, recordings, etc.)

### ⚠️ Cons

1. **Requires Training**: Need target speaker's audio (30+ minutes)
2. **Computational**: Needs GPU for training (CPU very slow)
3. **Quality Depends on Data**: More/better data = better results
4. **Not Zero-Shot**: Can't clone voice from single sample (unlike some newer systems)

---

## RVC vs Other Voice Conversion Methods

### RVC vs Traditional VC

**Traditional VC**:
- Statistical models (GMM, etc.)
- Requires more data
- Less natural results
- Older approach

**RVC**:
- Neural networks
- Better quality
- More data efficient
- Modern approach

### RVC vs Zero-Shot VC (OpenVoice, XTTS-v2)

**Zero-Shot VC**:
- ✅ No training needed
- ✅ Works with single sample
- ⚠️ May have lower quality
- ⚠️ Limited language support (XTTS: 17 languages)

**RVC**:
- ⚠️ Requires training (30+ minutes)
- ✅ Better quality with sufficient data
- ✅ Works with any language (HuBERT is cross-lingual)

---

## Summary

### What RVC Is

- **Voice cloning technique** (not TTS)
- **Speech-to-speech conversion** (audio in → audio out)
- **Neural approach** using modern AI (HuBERT, RMVPE, etc.)
- **Retrieval-based** (uses FAISS to find similar voice segments)

### What RVC Does

1. Extracts **what is said** (content features)
2. Extracts **how it's said** (F0/pitch)
3. Retrieves **target speaker characteristics** (FAISS)
4. Combines everything to **convert voice**
5. Generates **audio in target voice**

### What RVC Needs

- **Training data**: 30+ minutes of target speaker audio
- **GPU**: For training (CPU too slow)
- **Time**: 2-8 hours training time
- **Quality data**: Clean, consistent recordings

### Why It Might Work Better Than So-VITS-SVC

**FCBH RVC Claims**:
- Optimized for low-resource scenarios
- Works with less data (300-500 samples functional)
- Prosody-enhanced embeddings (30-45% improvement)
- Designed for minority languages

**However**: These are **unverified claims**. Worth testing to see if it actually works better with limited data.

---

## Next Steps

Given your situation:
1. **Test FCBH RVC** with same 26 segments
2. **Compare** to So-VITS-SVC results
3. **If better**: Use FCBH RVC
4. **If same/worse**: Collect more training data (30+ minutes)
5. **Alternative**: Use corpus snippets when available (no training needed)

---

## References

- FCBH RVC Code: `/Users/jrstear/git/FCBH-W2V-Bert-2.0-ASR-trainer/`
- VOICE_TRAINING_TAB_DOCUMENTATION.md: Comprehensive FCBH RVC docs
- RVC_EXPLANATION.md: Previous explanation document
- Wikipedia: Retrieval-based Voice Conversion

