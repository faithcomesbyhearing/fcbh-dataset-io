# Training Issue: fairseq + Python 3.11 Compatibility

**Issue**: `ValueError: mutable default <class 'fairseq.dataclass.configs.CommonConfig'> for field common is not allowed: use default_factory`

**Cause**: fairseq 0.12.2 has compatibility issues with Python 3.11's stricter dataclass handling.

**So-VITS-SVC Recommendation**: Python 3.8.9

## Solutions

### Option 1: Use Python 3.10 (Recommended)

Recreate conda environment with Python 3.10:

```bash
conda deactivate
conda remove --name revise_audio_vits --all -y
conda create --name revise_audio_vits python=3.10 -y
conda activate revise_audio_vits
conda install -y openblas
pip install --upgrade pip==24.0
cd $SO_VITS_SVC_ROOT
pip install -r requirements.txt
```

### Option 2: Patch fairseq (Quick Fix)

Edit the fairseq source to fix the dataclass issue:

```bash
# Find fairseq installation
python -c "import fairseq; print(fairseq.__file__)"

# Edit the problematic file (usually in dataclass/configs.py)
# Change mutable defaults to use default_factory
```

### Option 3: Use Different Encoder

Use an encoder that doesn't require fairseq (e.g., hubertsoft-onnx, whisper-ppg).

## Current Status

- ✅ Training data extracted
- ✅ Resampling complete
- ✅ File lists and config generated
- ❌ HuBERT preprocessing blocked by fairseq issue

## Next Steps

1. Recreate environment with Python 3.10
2. Re-run preprocessing steps
3. Start training

