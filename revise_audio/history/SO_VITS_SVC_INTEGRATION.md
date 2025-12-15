# So-VITS-SVC Integration Notes

**Date**: 2024-12-14  
**Status**: Initial setup - API research needed

## Overview

So-VITS-SVC (Retrieval-based Voice Conversion) will be integrated into the `revise_audio` module to provide neural voice conversion capabilities.

## Directory Structure

```
revise_audio/vits/
├── README.md
├── build_env.sh              # Conda environment setup
├── python/
│   └── so_vits_svc_inference.py  # Inference wrapper
├── models/                   # Trained model checkpoints (gitignored)
└── configs/                  # Model configurations
```

## Installation Options

### Option 1: Install from Source (Recommended)

```bash
cd revise_audio/vits
git clone https://github.com/svc-develop-team/so-vits-svc.git so_vits_svc_repo
cd so_vits_svc_repo
pip install -e .
```

### Option 2: Install as Package (if available)

```bash
pip install so-vits-svc
```

## API Research Needed

The actual So-VITS-SVC API needs to be researched. Key questions:

1. **Inference Interface**: How to load a model and convert audio?
   - Is there a Python API?
   - Or do we need to call CLI scripts?
   - What's the function signature?

2. **Model Format**: What does a trained model look like?
   - Directory structure?
   - Required files (checkpoint, config, etc.)?

3. **Input/Output**: 
   - Audio format requirements?
   - Sample rate handling?
   - F0 method options?

4. **Dependencies**: 
   - What Python packages are required?
   - Any system-level dependencies?

## Next Steps

1. ✅ Create directory structure
2. ✅ Create build_env.sh
3. ✅ Create Python inference wrapper (placeholder)
4. ✅ Create Go adapter
5. ⏳ Research actual So-VITS-SVC API
6. ⏳ Implement actual inference code
7. ⏳ Test with sample model
8. ⏳ Create training pipeline

## References

- So-VITS-SVC GitHub: https://github.com/svc-develop-team/so-vits-svc
- Documentation: Check repository README and wiki

