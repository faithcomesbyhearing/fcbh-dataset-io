# So-VITS-SVC Integration Decision

**Date**: 2024-12-14  
**Decision**: Keep So-VITS-SVC as a **separate external dependency**

## Analysis

### Repository Structure

So-VITS-SVC is a complete, self-contained project with:
- Training scripts (`train.py`, `train_diff.py`, `train_index.py`)
- Inference API (`inference/infer_tool.py` with `Svc` class)
- CLI interface (`inference_main.py`)
- Model definitions (`models.py`)
- Utilities and supporting code
- Configuration templates
- Web UI (`webUI.py`, `flask_api.py`)

### Key API

The main interface is the `Svc` class in `inference/infer_tool.py`:

```python
from inference.infer_tool import Svc

svc_model = Svc(
    net_g_path="logs/44k/G_37600.pth",      # Model checkpoint
    config_path="logs/44k/config.json",     # Config file
    device=None,                             # Auto-detect
    cluster_model_path="",                   # Optional
    nsf_hifigan_enhance=False,              # Optional enhancement
    diffusion_model_path="",                 # Optional diffusion
    diffusion_config_path="",                # Optional diffusion config
    shallow_diffusion=False,                 # Optional
    only_diffusion=False,                    # Optional
    spk_mix_enable=False,                    # Optional
    feature_retrieval=False                  # Optional
)

# Convert audio
audio = svc_model.slice_inference(
    raw_audio_path="input.wav",
    spk="speaker_name",
    tran=0,                                  # Pitch shift (semitones)
    slice_db=-40,                            # Slice threshold
    cluster_infer_ratio=0,                   # Cluster ratio
    auto_predict_f0=False,                   # Auto F0
    noice_scale=0.4,                         # Noise scale
    pad_seconds=0.5,                         # Padding
    clip_seconds=0,                          # Clip length
    lg_num=0,                                # Linear gradient
    lgr_num=0.75,                            # Linear gradient retain
    f0_predictor="pm",                       # F0 method: pm, rmvpe, crepe, etc.
    enhancer_adaptive_key=0,                 # Enhancer key
    cr_threshold=0.05,                       # F0 filter threshold
    k_step=100,                              # Diffusion steps
    use_spk_mix=False,                       # Speaker mix
    second_encoding=False,                   # Second encoding
    loudness_envelope_adjustment=1           # Loudness adjustment
)
```

## Recommendation: Separate Dependency

### Pros

1. **Separation of Concerns**: So-VITS-SVC is a complete, independent project. Keeping it separate maintains clear boundaries.

2. **Independent Updates**: Can update So-VITS-SVC without touching arti code. This is important as:
   - So-VITS-SVC is actively maintained
   - Updates may include bug fixes, new features
   - We may want to test different versions

3. **Matches Existing Pattern**: Similar to how MMS is handled - external tool, referenced via paths.

4. **Cleaner Repository**: Keeps arti focused on its core functionality.

5. **Version Control**: Can track So-VITS-SVC separately, potentially as a git submodule or just a separate clone.

6. **License Compliance**: So-VITS-SVC has its own license (AGPL 3.0). Keeping it separate makes license management clearer.

### Cons

1. **Path Management**: Need to configure paths to So-VITS-SVC
   - **Solution**: Use environment variable `SO_VITS_SVC_ROOT` or similar

2. **Setup Complexity**: Slightly more complex initial setup
   - **Solution**: Document in README and build scripts

## Implementation Approach

### 1. Path Configuration

Use environment variable to point to So-VITS-SVC:

```bash
export SO_VITS_SVC_ROOT=/Users/jrstear/git/so-vits-svc
```

### 2. Python Wrapper

Create a wrapper in `arti/revise_audio/vits/python/so_vits_svc_inference.py` that:
- Adds So-VITS-SVC to Python path
- Imports `Svc` class
- Provides a clean interface matching our needs
- Handles JSON I/O for Go communication

### 3. Integration Pattern

Follow the same pattern as MMS:
- Go code (`vits_adapter.go`) calls Python script
- Python script (`so_vits_svc_inference.py`) wraps So-VITS-SVC
- Communication via JSON over stdin/stdout

### 4. Dependencies

So-VITS-SVC has its own `requirements.txt`. We should:
- Install So-VITS-SVC dependencies in our conda environment
- Or document that So-VITS-SVC needs its own environment
- **Recommendation**: Install in same environment (simpler)

## Directory Structure

```
/Users/jrstear/git/
├── arti/                          # Our project
│   └── revise_audio/
│       └── vits/
│           ├── python/
│           │   └── so_vits_svc_inference.py  # Wrapper
│           └── ...
└── so-vits-svc/                   # External dependency
    ├── inference/
    │   └── infer_tool.py          # Svc class
    ├── models.py
    └── ...
```

## Next Steps

1. ✅ Update `so_vits_svc_inference.py` to:
   - Add So-VITS-SVC to Python path via `SO_VITS_SVC_ROOT`
   - Import and use `Svc` class
   - Implement actual inference

2. Update `build_env.sh` to:
   - Install So-VITS-SVC dependencies
   - Note requirement for `SO_VITS_SVC_ROOT` environment variable

3. Update documentation to explain setup

4. Test with a sample model

