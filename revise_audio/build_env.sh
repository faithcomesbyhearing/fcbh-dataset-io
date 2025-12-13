#!/bin/bash

# Build environment for revise_audio Python dependencies
# Supports both M1 Mac (development) and AWS EC2 g6e.xlarge (production)

# This is meant to be executed line by line, or as a script

conda deactivate
conda remove --name revise_audio --all -y

conda create --name revise_audio python=3.11 -y
conda activate revise_audio
pip install --upgrade pip
pip install numpy

# PyTorch installation - platform specific
if [ "$(uname)" == "Darwin" ]; then
  # M1 Mac - CPU only (MPS support if needed)
  pip install torch torchaudio
else
  # Linux (AWS EC2) - CUDA support
  pip install torch torchaudio --index-url https://download.pytorch.org/whl/cu121
fi

# Core audio processing
pip install soundfile
pip install librosa>=0.10.0
pip install scipy>=1.9.0

# Voice conversion dependencies
pip install transformers>=4.25.0  # HuBERT
pip install speechbrain>=0.5.15   # X-Vector embeddings

# F0 extraction
pip install pyworld>=0.3.2
# Note: pyworld may require system libraries on Linux:
#   sudo apt-get install build-essential
#   sudo apt-get install libsndfile1-dev

# Optional: torchcrepe for alternative F0 extraction
# pip install torchcrepe>=0.0.20

# FAISS for similarity search (if needed for RVC)
if [ "$(uname)" == "Darwin" ]; then
  pip install faiss-cpu
else
  pip install faiss-gpu  # Requires CUDA
fi

# Error handling (from logger package)
# This assumes the logger package is in the repo
# If not, remove this dependency

conda deactivate

echo ""
echo "Environment 'revise_audio' created successfully!"
echo ""
echo "Set environment variable in your shell profile:"
echo "  export FCBH_REVISE_AUDIO_VC_PYTHON=\$CONDA_PREFIX/bin/python"
echo "  export FCBH_REVISE_AUDIO_PROSODY_PYTHON=\$CONDA_PREFIX/bin/python"
echo ""
echo "Or for conda environment:"
echo "  export FCBH_REVISE_AUDIO_VC_PYTHON=\$CONDA_PREFIX/envs/revise_audio/bin/python"
echo "  export FCBH_REVISE_AUDIO_PROSODY_PYTHON=\$CONDA_PREFIX/envs/revise_audio/bin/python"

