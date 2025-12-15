# Session Summary: So-VITS-SVC Integration

**Date**: 2024-12-14  
**Session Goal**: Integrate So-VITS-SVC for neural voice conversion and prepare for testing on Jude 1:1

---

## Major Accomplishments

### 1. So-VITS-SVC Integration Structure ✅

Created complete integration structure in `revise_audio/vits/`:
- **Python Wrapper**: `python/so_vits_svc_inference.py` - Wraps So-VITS-SVC `Svc` class
- **Go Adapter**: `vits_adapter.go` - Provides Go interface following MMS pattern
- **Build Script**: `build_env.sh` - Conda environment setup with dependency management
- **Test Scripts**: Import test and training data extraction scripts
- **Documentation**: README, setup instructions, integration notes

### 2. Environment Setup ✅

- **Decision**: Keep So-VITS-SVC as separate external dependency (like MMS)
- **Environment Variable**: `SO_VITS_SVC_ROOT` points to cloned repository
- **Conda Environment**: `revise_audio_vits` with Python 3.10
- **Dependencies**: All So-VITS-SVC requirements installed successfully

### 3. Training Data Preparation ✅

- **Extracted**: 26 verse segments from Jude chapter 1
- **Location**: `$SO_VITS_SVC_ROOT/dataset_raw/jude_narrator/`
- **Format**: WAV files, 44.1kHz mono (resampled)
- **Preprocessing**: HuBERT features and F0 extracted successfully

### 4. Compatibility Issues Resolved ✅

**Issue 1**: Python 3.11 + fairseq compatibility
- **Solution**: Recreated environment with Python 3.10

**Issue 2**: PyTorch 2.6+ `weights_only` default
- **Solution**: Downgraded to PyTorch 2.5.1

**Issue 3**: RMVPE model missing
- **Solution**: Used PM F0 predictor instead

---

## Key Decisions

1. **External Dependency**: So-VITS-SVC kept separate (not incorporated into arti)
   - Matches existing pattern (MMS)
   - Easier to update independently
   - Clear separation of concerns

2. **Python Version**: 3.10 (not 3.8.9 as recommended, but compatible)
   - Better compatibility with modern packages
   - Resolves fairseq issues

3. **PyTorch Version**: 2.5.1 (downgraded from 2.9.1)
   - Avoids `weights_only` security changes
   - Compatible with fairseq checkpoint loading

---

## Files Created

### Integration Code
- `revise_audio/vits_adapter.go` - Go adapter for So-VITS-SVC
- `revise_audio/vits/python/so_vits_svc_inference.py` - Python inference wrapper
- `revise_audio/vits/python/test_import.py` - Import verification script
- `revise_audio/vits/python/prepare_training_data.py` - Training data extraction
- `revise_audio/vits/python/train_model.sh` - Training pipeline script
- `revise_audio/vits/build_env.sh` - Environment setup script

### Test Scripts
- `revise_audio/cmd/test_vits_import/main.go` - Go test for import
- `revise_audio/cmd/test_vits_jude/main.go` - Full workflow test for Jude 1:1

### Documentation
- `revise_audio/vits/README.md` - Integration overview
- `revise_audio/vits/SETUP_INSTRUCTIONS.md` - Setup guide
- `revise_audio/vits/INTEGRATION_STATUS.md` - Status tracking
- `revise_audio/vits/SETUP_COMPLETE.md` - Setup completion notes
- `revise_audio/vits/TRAINING_STATUS.md` - Training preparation
- `revise_audio/vits/TRAINING_ISSUE.md` - Compatibility issues
- `revise_audio/vits/TRAINING_READY.md` - Training readiness
- `revise_audio/history/SO_VITS_SVC_INTEGRATION.md` - Integration notes
- `revise_audio/history/SO_VITS_SVC_INTEGRATION_DECISION.md` - Architecture decision

### Modified Files
- `revise_audio/vits/build_env.sh` - Updated for Python 3.10 and openblas
- `revise_audio/models.go` - Already had VoiceConversionConfig (no changes needed)

---

## Current Status

### ✅ Complete
- So-VITS-SVC integration code
- Environment setup and dependency installation
- Training data extraction and preprocessing
- Import verification
- Test scripts ready

### ⏳ Pending
- **GPU Training**: Requires AWS EC2 g6e-xlarge (or similar GPU instance)
- **Model Training**: Run training pipeline to create speaker model
- **Inference Testing**: Test voice conversion on Jude 1:1 after training

---

## Next Steps

### Immediate (On AWS EC2 GPU)

1. **Transfer Files**:
   ```bash
   # Copy preprocessing results to AWS EC2
   scp -r $SO_VITS_SVC_ROOT/dataset/44k user@ec2:~/so-vits-svc/dataset/
   scp -r $SO_VITS_SVC_ROOT/configs user@ec2:~/so-vits-svc/
   scp -r $SO_VITS_SVC_ROOT/filelists user@ec2:~/so-vits-svc/
   ```

2. **Train Model**:
   ```bash
   cd $SO_VITS_SVC_ROOT
   python train.py -c configs/config.json
   ```

3. **Monitor Training**:
   ```bash
   tensorboard --logdir logs/44k
   ```

### After Training

4. **Test Inference**:
   ```bash
   cd arti
   go run revise_audio/cmd/test_vits_jude/main.go
   ```

5. **Integrate into Workflow**:
   - Update `test_tts_segments/main.go` to use `VITSAdapter` instead of `RVCAdapter`
   - Test full pipeline: TTS → VC → Prosody → Stitch

---

## Technical Notes

### API Usage

**Python (Direct)**:
```python
from inference.infer_tool import Svc

svc_model = Svc(
    net_g_path="logs/44k/G_37600.pth",
    config_path="logs/44k/config.json",
    device=None
)

audio = svc_model.slice_inference(
    raw_audio_path="input.wav",
    spk="jude_narrator",
    tran=0,
    f0_predictor="rmvpe",
    # ... other params
)
```

**Go (via Adapter)**:
```go
adapter := revise_audio.NewVITSAdapter(ctx, config)
convertedPath, status := adapter.ConvertVoice(
    sourceAudioPath,
    modelPath,      // Path to .pth checkpoint
    configPath,     // Path to config.json
    speaker,        // Speaker name from model
)
```

### Training Data

- **Source**: Jude chapter 1, all 26 verses
- **Duration**: ~4 minutes total audio
- **Format**: 44.1kHz, mono WAV
- **Speaker**: "jude_narrator" (single speaker model)

### Configuration

- **Sample Rate**: 44.1kHz
- **F0 Predictor**: PM (for preprocessing), RMVPE (for inference, if available)
- **Speech Encoder**: ContentVec768L12 (vec768l12)
- **Vocoder**: NSF-HiFiGAN (if available)

---

## Issues Encountered & Resolved

1. **fairseq + Python 3.11**: Incompatible dataclass handling
   - **Fix**: Python 3.10

2. **PyTorch 2.6+ weights_only**: Security change breaks checkpoint loading
   - **Fix**: PyTorch 2.5.1

3. **RMVPE model missing**: F0 predictor not available
   - **Fix**: Use PM predictor for preprocessing

4. **NumPy/OpenBLAS on macOS**: Library loading issues
   - **Fix**: Install openblas via conda

5. **Disk space**: Installation blocked by low disk space
   - **Fix**: Cleaned conda/pip caches

---

## Beads Status

- **arti-2ps**: In progress - So-VITS-SVC integration
  - Integration code complete ✅
  - Environment setup complete ✅
  - Training data ready ✅
  - Pending: GPU training and inference testing

---

## Lessons Learned

1. **Version Compatibility**: So-VITS-SVC has specific version requirements
   - Python 3.8.9 recommended, but 3.10 works
   - PyTorch < 2.6 required for fairseq compatibility
   - fairseq 0.12.2 has strict requirements

2. **External Dependencies**: Keeping So-VITS-SVC separate is the right approach
   - Easier maintenance
   - Clear boundaries
   - Matches existing patterns

3. **Training Requirements**: GPU is mandatory for training
   - CPU training explicitly disabled
   - Need AWS EC2 or similar for actual training

4. **Preprocessing**: Can be done on CPU (M1 Mac)
   - Data extraction ✅
   - Resampling ✅
   - Feature extraction ✅
   - Only training requires GPU

---

## References

- So-VITS-SVC GitHub: https://github.com/svc-develop-team/so-vits-svc
- Integration Decision: `revise_audio/history/SO_VITS_SVC_INTEGRATION_DECISION.md`
- Training Status: `revise_audio/vits/TRAINING_READY.md`
- Setup Instructions: `revise_audio/vits/SETUP_INSTRUCTIONS.md`

---

## Session Metrics

- **Files Created**: 15+
- **Issues Resolved**: 5 compatibility issues
- **Training Data**: 26 audio segments extracted
- **Preprocessing**: Complete
- **Integration**: Code complete, pending training

---

**Next Session**: Train model on AWS EC2, then test inference on Jude 1:1

