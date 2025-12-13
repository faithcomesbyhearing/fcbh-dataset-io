# Audio Revision System

Audio find-and-replace system for verse-level corrections in Bible recordings.

## Overview

This module provides functionality to:
1. Find replacement words/phrases in existing corpus
2. Extract audio snippets
3. Apply voice conversion (if different speaker)
4. Match prosody to surrounding context
5. Stitch revised audio into chapter files

## Directory Structure

```
revise_audio/
├── models.go              # Core data structures
├── revise_audio.go        # Main module
├── corpus_search.go       # Corpus search for replacement snippets
├── vc_adapter.go          # Voice conversion interface & RVC adapter
├── prosody_adapter.go     # Prosody matching interface & DSP adapter
├── build_env.sh           # Conda environment setup script
├── README.md              # This file
└── python/
    ├── voice_conversion.py  # RVC Python module
    └── prosody_match.py     # Prosody matching Python module
```

## Setup

### 1. Create Conda Environment

```bash
cd revise_audio
./build_env.sh
```

Or execute line by line for better error handling.

### 2. Set Environment Variables

Add to your shell profile (`.bashrc`, `.zshrc`, etc.):

```bash
# For conda environment
export FCBH_REVISE_AUDIO_VC_PYTHON=$CONDA_PREFIX/envs/revise_audio/bin/python
export FCBH_REVISE_AUDIO_PROSODY_PYTHON=$CONDA_PREFIX/envs/revise_audio/bin/python

# Or if using system Python
export FCBH_REVISE_AUDIO_VC_PYTHON=/path/to/python
export FCBH_REVISE_AUDIO_PROSODY_PYTHON=/path/to/python
```

### 3. Platform Notes

**M1 Mac (Development)**:
- Uses CPU-only PyTorch
- FAISS CPU version
- May need Rosetta for some packages

**AWS EC2 g6e.xlarge (Production)**:
- CUDA-enabled PyTorch
- FAISS GPU version
- Requires CUDA 12.1+ and cuDNN

## Usage

### Corpus Search (`corpus_search.go`)

The `CorpusSearcher` provides functionality to find replacement audio snippets in the corpus:

```go
searcher := NewCorpusSearcher(ctx, dbAdapter)
candidates, status := searcher.FindReplacementSnippets(
    "replacement text",
    "MAT",  // target book
    1,      // target chapter
    5,      // target verse
    false,  // returnAll: false = best match only, true = all ranked candidates
)
```

**Features:**
- Exact text matching (case-insensitive, normalized)
- Phrase matching: finds consecutive word sequences
- Distance-based ranking: prioritizes nearby verses
- Returns `SnippetCandidate` with metadata (book, chapter, verse, timestamps, actor, person)

See individual task implementations:
- `arti-44l`: Corpus search ✅ (implemented)
- `arti-7qf`: Snippet extraction
- `arti-lmh`: Voice conversion
- `arti-a2g`: Prosody matching
- `arti-bue`: Audio stitching

## Dependencies

See `TOOLING_RESEARCH.md` for detailed dependency information.

## Integration

This module integrates with:
- `db` package: Script and Word tables
- `input` package: Audio file access
- `output` package: Revised audio output
- `controller`: Main workflow orchestration

