#!/bin/bash

# Build environment for So-VITS-SVC
# Supports both M1 Mac (development) and AWS EC2 g6e.xlarge (production)

# This is meant to be executed line by line, or as a script

# Initialize conda if needed
if [ -z "$CONDA_DEFAULT_ENV" ]; then
    # Try to source conda.sh if available
    if [ -f "$(conda info --base 2>/dev/null)/etc/profile.d/conda.sh" ]; then
        source "$(conda info --base)/etc/profile.d/conda.sh"
    fi
fi

# Remove existing environment if it exists
if conda env list | grep -q "revise_audio_vits"; then
    conda remove --name revise_audio_vits --all -y 2>/dev/null || true
fi

conda create --name revise_audio_vits python=3.11 -y

# Activate conda environment
if [ -f "$(conda info --base 2>/dev/null)/etc/profile.d/conda.sh" ]; then
    source "$(conda info --base)/etc/profile.d/conda.sh"
fi
conda activate revise_audio_vits

# Install openblas for numpy (required on macOS)
conda install -y openblas

# Downgrade pip to 24.0 to handle fairseq dependency conflicts
pip install --upgrade pip==24.0
pip install numpy

# PyTorch installation - platform specific
if [ "$(uname)" == "Darwin" ]; then
  # M1 Mac - CPU only (MPS support if needed)
  pip install torch torchaudio
else
  # Linux (AWS EC2) - CUDA support
  pip install torch torchaudio --index-url https://download.pytorch.org/whl/cu121
fi

# So-VITS-SVC dependencies
# Install from So-VITS-SVC requirements.txt if available
SO_VITS_SVC_ROOT="${SO_VITS_SVC_ROOT:-}"
if [ -n "$SO_VITS_SVC_ROOT" ] && [ -f "$SO_VITS_SVC_ROOT/requirements.txt" ]; then
  echo "Installing So-VITS-SVC dependencies from $SO_VITS_SVC_ROOT/requirements.txt..."
  pip install -r "$SO_VITS_SVC_ROOT/requirements.txt"
else
  echo "Warning: SO_VITS_SVC_ROOT not set or requirements.txt not found."
  echo "Installing core dependencies manually..."
  
  # Core audio processing (matching So-VITS-SVC requirements)
  pip install soundfile==0.12.1
  pip install librosa==0.9.1
  pip install scipy==1.10.0
  pip install numpy==1.23.5
  
  # So-VITS-SVC core dependencies
  pip install transformers
  pip install pyworld
  pip install torchcrepe
  pip install fairseq==0.12.2
  pip install onnx onnxsim onnxoptimizer
  pip install tqdm rich loguru
  pip install scikit-maad praat-parselmouth
  pip install tensorboard tensorboardX
  pip install edge_tts langdetect pyyaml pynvml
  pip install einops local_attention
  pip install ffmpeg-python Flask Flask_Cors gradio>=3.7.0
  
  # FAISS for similarity search
  if [ "$(uname)" == "Darwin" ]; then
    pip install faiss-cpu
  else
    pip install faiss-gpu  # Requires CUDA
  fi
fi

# So-VITS-SVC is expected to be cloned separately
# Set SO_VITS_SVC_ROOT environment variable to point to the clone location
# Example: export SO_VITS_SVC_ROOT=/Users/jrstear/git/so-vits-svc
echo ""
echo "📦 So-VITS-SVC Setup:"
echo ""
if [ -n "$SO_VITS_SVC_ROOT" ] && [ -f "$SO_VITS_SVC_ROOT/requirements.txt" ]; then
  echo "   ✓ SO_VITS_SVC_ROOT is set: $SO_VITS_SVC_ROOT"
  echo "   ✓ Dependencies installed from requirements.txt"
else
  echo "   ⚠️  So-VITS-SVC should be cloned separately:"
  echo "      git clone https://github.com/svc-develop-team/so-vits-svc.git"
  echo ""
  echo "   Then set environment variable before running this script:"
  echo "      export SO_VITS_SVC_ROOT=/path/to/so-vits-svc"
  echo "      bash build_env.sh"
  echo ""
  echo "   Or install dependencies manually:"
  echo "      cd \$SO_VITS_SVC_ROOT"
  echo "      pip install -r requirements.txt"
fi
echo ""
echo "   The Python wrapper will add So-VITS-SVC to sys.path automatically."
echo ""

conda deactivate

echo ""
echo "Environment 'revise_audio_vits' created successfully!"
echo ""
echo "Set environment variable in your shell profile:"
echo "  export FCBH_VITS_PYTHON=\$CONDA_PREFIX/envs/revise_audio_vits/bin/python"
echo ""

