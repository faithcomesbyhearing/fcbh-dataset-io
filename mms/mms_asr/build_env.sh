#!/bin/bash

conda create -y -n mms_asr python=3.11

conda activate mms_asr

if [ "$(uname)" == "Darwin" ]; then
  pip install torch torchaudio
else
  pip install torch torchaudio --index-url https://download.pytorch.org/whl/cpu
fi

# conda install -y pytorch torchaudio pytorch-cuda=12.1 -c pytorch -c nvidia

# On Mac
# conda install -y pytorch::pytorch torchaudio -c pytorch

pip install accelerate
pip install datasets
pip install --upgrade transformers
pip install soundfile
pip install librosa

pip install uroman
if [ "$(uname)" == "Darwin" ]; then
  cp $HOME/miniforge/envs/mms_fa/bin/uroman $HOME/miniforge/envs/mms_fa/bin/uroman.pl
else
  cp /opt/conda/envs/mms_fa/bin/uroman /opt/conda/envs/mms_fa/bin/uroman.pl
fi
