# So-VITS-SVC Training Status

**Date**: 2024-12-14  
**Status**: Training data ready, ready to train

## Training Data

✅ **Extracted**: 26 verse segments from Jude chapter 1  
✅ **Location**: `$SO_VITS_SVC_ROOT/dataset_raw/jude_narrator/`  
✅ **Format**: WAV files, extracted from original recordings

## Next Steps

### 1. Download Required Pretrained Models

**HuBERT (ContentVec) - Required:**
```bash
cd $SO_VITS_SVC_ROOT/pretrain
wget https://huggingface.co/lj1995/VoiceConversionWebUI/resolve/main/hubert_base.pt -O checkpoint_best_legacy_500.pt
```

Or download from: https://ibm.box.com/s/z1wgl1stco8ffooyatzdwsqn2psd9lrr

**RMVPE (F0 Predictor) - Recommended:**
```bash
cd $SO_VITS_SVC_ROOT/pretrain
wget https://github.com/yxlllc/RMVPE/releases/download/230917/rmvpe.zip
unzip rmvpe.zip
mv model.pt rmvpe.pt
```

### 2. Run Training Pipeline

```bash
export SO_VITS_SVC_ROOT=/Users/jrstear/git/so-vits-svc
source $(conda info --base)/etc/profile.d/conda.sh
conda activate revise_audio_vits
cd $SO_VITS_SVC_ROOT
bash ../arti/revise_audio/vits/python/train_model.sh
```

Or run steps manually:
```bash
cd $SO_VITS_SVC_ROOT
python resample.py
python preprocess_flist_config.py
python preprocess_hubert_f0.py
python train.py -c configs/config.json
```

### 3. Monitor Training

```bash
tensorboard --logdir logs/44k
```

Training typically takes several hours depending on:
- Amount of training data (we have 26 segments, ~4 minutes total)
- GPU availability (CPU training is very slow)
- Number of epochs

### 4. Test Inference

Once training completes (checkpoints in `logs/44k/G_*.pth`), test with:
```bash
cd arti
go run revise_audio/cmd/test_vits_jude/main.go
```

## Notes

- Training on CPU will be very slow (hours/days)
- For faster training, use GPU (AWS EC2 g6e.xlarge)
- Model checkpoints are saved periodically during training
- Can test inference with partial training (early checkpoints)

