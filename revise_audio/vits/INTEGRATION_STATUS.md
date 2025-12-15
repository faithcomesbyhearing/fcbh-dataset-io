# So-VITS-SVC Integration Status

**Date**: 2024-12-14  
**Status**: Implementation complete, ready for testing

## Completed

✅ **Directory Structure**: Created `revise_audio/vits/` with proper organization  
✅ **Python Wrapper**: `python/so_vits_svc_inference.py` - Wraps So-VITS-SVC `Svc` class  
✅ **Go Adapter**: `vits_adapter.go` - Provides Go interface for voice conversion  
✅ **Build Script**: `build_env.sh` - Conda environment setup  
✅ **Documentation**: README and integration decision docs  
✅ **Test Scripts**: Import test to verify setup  

## Setup Required

### 1. Set Environment Variable

```bash
export SO_VITS_SVC_ROOT=/Users/jrstear/git/so-vits-svc
```

Add to your shell profile (`~/.zshrc` or `~/.bashrc`) for persistence.

### 2. Install So-VITS-SVC Dependencies

```bash
cd $SO_VITS_SVC_ROOT
pip install -r requirements.txt
```

Or install in the conda environment:

```bash
conda activate revise_audio_vits  # or revise_audio
cd $SO_VITS_SVC_ROOT
pip install -r requirements.txt
```

### 3. Install arti Dependencies

```bash
cd arti/revise_audio/vits
bash build_env.sh
```

### 4. Verify Setup

```bash
# Test import
export SO_VITS_SVC_ROOT=/Users/jrstear/git/so-vits-svc
python3 revise_audio/vits/python/test_import.py

# Or use Go test
go run revise_audio/cmd/test_vits_import/main.go
```

## API Usage

### Python (Direct)

```python
from inference.infer_tool import Svc

svc_model = Svc(
    net_g_path="logs/44k/G_37600.pth",
    config_path="logs/44k/config.json",
    device=None
)

audio = svc_model.slice_inference(
    raw_audio_path="input.wav",
    spk="speaker0",
    tran=0,
    slice_db=-40,
    cluster_infer_ratio=0,
    auto_predict_f0=False,
    noice_scale=0.4,
    pad_seconds=0.5,
    clip_seconds=0,
    lg_num=0,
    lgr_num=0.75,
    f0_predictor="rmvpe",
    enhancer_adaptive_key=0,
    cr_threshold=0.05,
    k_step=100,
    use_spk_mix=False,
    second_encoding=False,
    loudness_envelope_adjustment=1
)
```

### Go (via Adapter)

```go
adapter := revise_audio.NewVITSAdapter(ctx, config)
convertedPath, status := adapter.ConvertVoice(
    sourceAudioPath,
    modelPath,      // Path to .pth checkpoint
    configPath,     // Path to config.json
    speaker,        // Speaker name from model
)
```

## Next Steps

1. ⏳ **Test Import**: Verify So-VITS-SVC can be imported
2. ⏳ **Train Test Model**: Train a speaker model on test data (Jude narrator)
3. ⏳ **Integration Test**: Test full voice conversion workflow
4. ⏳ **Update vc_adapter.go**: Replace DSP-based VC with So-VITS-SVC
5. ⏳ **End-to-End Test**: Test with actual revision workflow

## Known Issues

- None yet - awaiting testing

## Dependencies

So-VITS-SVC requires:
- PyTorch
- librosa
- soundfile
- numpy
- scipy
- transformers (for HuBERT)
- pyworld / torchcrepe (for F0)
- fairseq (for HuBERT models)
- See `$SO_VITS_SVC_ROOT/requirements.txt` for complete list

