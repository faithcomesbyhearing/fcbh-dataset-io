#!/bin/bash

# This is meant to be executed line by line

conda activate base
#conda remove --name mms_adapter --all

conda create -y -n mms_adapter python=3.11
conda activate mms_adapter
pip install --upgrade pip
pip install numpy
if [ "$(uname)" == "Darwin" ]; then
  pip install torch torchaudio
else
  pip install torch torchaudio --index-url https://download.pytorch.org/whl/cpu
fi
pip install soundfile
pip install adapters
pip install peft
pip install jiwer
pip install evaluate
#pip install unicodedata
#pip install accelerate

## Required environment variable in .bash_profile
## export FCBH_MMS_ADAPTER_PYTHON=$PY_ENV/mms_adapter/bin/python

