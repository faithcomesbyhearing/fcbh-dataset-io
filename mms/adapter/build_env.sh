#!/bin/bash

# This is meant to be executed line by line

conda deactivate
conda remove --name mms_adapter --all

conda create --name mms_adapter python=3.11 -y
conda activate mms_adapter
pip install --upgrade pip
pip install numpy
if [ "$(uname)" == "Darwin" ]; then
  pip install torch torchaudio
else
  pip install torch torchaudio --index-url https://download.pytorch.org/whl/cu118
fi
pip install soundfile
pip install adapters
pip install peft
pip install jiwer
pip install evaluate
#pip install unicodedata
#pip install accelerate

