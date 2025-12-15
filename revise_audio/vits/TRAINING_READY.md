# So-VITS-SVC Training Ready

**Date**: 2024-12-14  
**Status**: Preprocessing complete, ready for GPU training

## Completed Steps

✅ **Training Data Extracted**: 26 verse segments from Jude chapter 1  
✅ **HuBERT Model Downloaded**: ContentVec checkpoint ready  
✅ **Resampling Complete**: Audio resampled to 44.1kHz mono  
✅ **Config Generated**: File lists and config.json created  
✅ **Preprocessing Complete**: HuBERT features and F0 extracted  

## Environment Setup

- **Python**: 3.10 (fixed compatibility issues)
- **PyTorch**: 2.5.1 (downgraded from 2.9.1 to avoid weights_only issue)
- **F0 Predictor**: PM (used instead of RMVPE to avoid missing model)

## Training Location

Training requires GPU. The script enforces this:
```python
assert torch.cuda.is_available(), "CPU training is not allowed."
```

**Next**: Train on AWS EC2 g6e-xlarge (GPU instance)

## Training Command

On AWS EC2 with GPU:
```bash
export SO_VITS_SVC_ROOT=/path/to/so-vits-svc
source $(conda info --base)/etc/profile.d/conda.sh
conda activate revise_audio_vits
cd $SO_VITS_SVC_ROOT
python train.py -c configs/config.json
```

Monitor training:
```bash
tensorboard --logdir logs/44k
```

## After Training

Once training completes (checkpoints in `logs/44k/G_*.pth`), test inference:
```bash
cd arti
go run revise_audio/cmd/test_vits_jude/main.go
```

## Files Ready

- Training data: `dataset_raw/jude_narrator/` (26 WAV files)
- Processed data: `dataset/44k/jude_narrator/` (with .soft.pt and .f0.npy)
- Config: `configs/config.json`
- File lists: `filelists/train.txt`, `filelists/val.txt`

