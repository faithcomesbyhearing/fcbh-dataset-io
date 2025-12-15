# So-VITS-SVC Setup Instructions

## Quick Start

### 1. Set Environment Variable

```bash
export SO_VITS_SVC_ROOT=/Users/jrstear/git/so-vits-svc
```

Add to your shell profile for persistence:
```bash
echo 'export SO_VITS_SVC_ROOT=/Users/jrstear/git/so-vits-svc' >> ~/.zshrc
source ~/.zshrc
```

### 2. Install Dependencies

**Recommended: Use build script (automatically installs So-VITS-SVC dependencies)**

```bash
export SO_VITS_SVC_ROOT=/Users/jrstear/git/so-vits-svc
cd arti/revise_audio/vits
bash build_env.sh
conda activate revise_audio_vits
```

The build script will automatically detect `SO_VITS_SVC_ROOT` and install dependencies from `$SO_VITS_SVC_ROOT/requirements.txt`.

**Alternative: Install manually**

```bash
conda activate revise_audio_vits  # or revise_audio
cd $SO_VITS_SVC_ROOT
pip install -r requirements.txt
```

### 3. Verify Installation

```bash
export SO_VITS_SVC_ROOT=/Users/jrstear/git/so-vits-svc
python3 revise_audio/vits/python/test_import.py
```

Expected output:
```
✓ SO_VITS_SVC_ROOT: /Users/jrstear/git/so-vits-svc
✓ Path exists: True
Attempting to import So-VITS-SVC...
✓ Successfully imported Svc class from inference.infer_tool
✓ Svc.slice_inference method found
✓ Svc.__init__ method found
✓ So-VITS-SVC import test PASSED
```

## Troubleshooting

### "No module named 'librosa'"

Install So-VITS-SVC dependencies:
```bash
cd $SO_VITS_SVC_ROOT
pip install -r requirements.txt
```

### "SO_VITS_SVC_ROOT environment variable not set"

Set the environment variable:
```bash
export SO_VITS_SVC_ROOT=/Users/jrstear/git/so-vits-svc
```

### "SO_VITS_SVC_ROOT path does not exist"

Verify the path is correct:
```bash
ls -la $SO_VITS_SVC_ROOT/inference/infer_tool.py
```

### Import errors with specific modules

So-VITS-SVC has specific version requirements. Install from their requirements.txt:
```bash
cd $SO_VITS_SVC_ROOT
pip install -r requirements.txt
```

## Next Steps

Once setup is verified:

1. **Train a model** (see training documentation)
2. **Test inference** with a trained model
3. **Integrate into workflow** (update `vc_adapter.go` to use `VITSAdapter`)

