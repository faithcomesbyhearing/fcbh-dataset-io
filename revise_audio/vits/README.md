# So-VITS-SVC Integration

This directory contains the integration with So-VITS-SVC (Retrieval-based Voice Conversion) for neural voice conversion.

## Overview

So-VITS-SVC is used to convert audio from one speaker to sound like another speaker while preserving linguistic content. It supports both:
- **Text-driven revisions**: TTS output → voice conversion
- **Audio-only revisions**: Recorded snippets → voice conversion

## Directory Structure

```
vits/
├── README.md              # This file
├── build_env.sh           # Conda environment setup for So-VITS-SVC
├── python/
│   ├── so_vits_svc_inference.py  # Inference wrapper (converts audio)
│   └── so_vits_svc_train.py      # Training script (trains speaker models)
├── models/                # Trained model checkpoints (gitignored)
│   └── .gitkeep
└── configs/               # Model configurations (optional)
    └── .gitkeep
```

## Architecture

### Inference Flow

```
Source Audio (TTS or Recorded)
    ↓
So-VITS-SVC Inference
    ↓
[Extract content features (HuBERT)]
[Extract F0 (pitch)]
[Load target speaker model]
[Convert voice]
    ↓
Converted Audio (target speaker's voice)
```

### Training Flow

```
Speaker Corpus Audio Files
    ↓
Feature Extraction (HuBERT, F0, speaker embeddings)
    ↓
Model Training
    ↓
Trained Model Checkpoint
```

## Integration Pattern

Follows the same Go-Python pattern as other modules:
- **Go**: `vits_adapter.go` - Orchestration, file I/O, database access
- **Python**: `python/so_vits_svc_inference.py` - Actual So-VITS-SVC calls
- **Communication**: JSON over stdin/stdout (via `stdio_exec`)

## Dependencies

So-VITS-SVC is kept as a **separate external dependency** (similar to MMS).

### Setup

1. **Clone So-VITS-SVC separately**:
   ```bash
   cd /path/to/your/git/directory
   git clone https://github.com/svc-develop-team/so-vits-svc.git
   ```

2. **Set environment variable**:
   ```bash
   export SO_VITS_SVC_ROOT=/path/to/so-vits-svc
   ```

3. **Install So-VITS-SVC dependencies**:
   ```bash
   cd $SO_VITS_SVC_ROOT
   pip install -r requirements.txt
   ```

4. **Install arti dependencies** (in conda environment):
   ```bash
   cd arti/revise_audio/vits
   bash build_env.sh
   ```

The Python wrapper (`so_vits_svc_inference.py`) will automatically add So-VITS-SVC to the Python path.

See `build_env.sh` for complete dependency list.

## Usage

### Training a Speaker Model

```bash
# Activate environment
conda activate revise_audio_vits

# Train model
python python/so_vits_svc_train.py \
    --audio_dir /path/to/speaker/audio \
    --output_dir models/speaker_name \
    --config configs/default.json
```

### Using for Voice Conversion

```go
// In Go code
adapter := revise_audio.NewVITSAdapter(ctx, config)
convertedPath, status := adapter.ConvertVoice(
    sourceAudioPath,
    targetSpeakerModelPath,
    outputPath,
)
```

## References

- So-VITS-SVC GitHub: https://github.com/svc-develop-team/so-vits-svc
- Documentation: See `history/SO_VITS_SVC_INTEGRATION.md`

