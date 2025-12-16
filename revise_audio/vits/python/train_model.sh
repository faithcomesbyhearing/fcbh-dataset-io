#!/bin/bash
# So-VITS-SVC Training Script
# Prepares data and trains a model for the Jude narrator

set -e

SO_VITS_SVC_ROOT="${SO_VITS_SVC_ROOT:-}"
if [ -z "$SO_VITS_SVC_ROOT" ]; then
    echo "Error: SO_VITS_SVC_ROOT not set"
    exit 1
fi

cd "$SO_VITS_SVC_ROOT"

echo "=== So-VITS-SVC Training Pipeline ==="
echo ""

# Step 1: Resample audio to 44.1kHz mono
echo "Step 1: Resampling audio to 44.1kHz mono..."
python resample.py
echo ""

# Step 2: Generate file lists and config
echo "Step 2: Generating file lists and config..."
python preprocess_flist_config.py
echo ""

# Step 3: Preprocess HuBERT features and F0
echo "Step 3: Preprocessing HuBERT features and F0..."
python preprocess_hubert_f0.py
echo ""

# Step 4: Train model
echo "Step 4: Starting training..."
echo "Note: Training will take a while. Monitor progress with tensorboard:"
echo "  tensorboard --logdir logs/44k"
echo ""
python train.py -c configs/config.json -m 44k

echo ""
echo "Training complete! Check logs/44k/ for model checkpoints."

