#!/bin/bash

conda install -y pytorch torchaudio pytorch-cuda=12.1 -c pytorch -c nvidia

# On Mac
# conda install -y pytorch::pytorch torchaudio -c pytorch

pip install accelerate
pip install datasets
pip install --upgrade transformers
pip install soundfile
pip install librosa

pip install uroman
cp /opt/conda/envs/mms_asr/bin/uroman /opt/conda/envs/mms_asr/bin/uroman.pl
# on Mac
# cp /Users/gary/miniforge3/envs/mms_asr/bin/uroman /Users/gary/miniforge3/envs/mms_asr/bin/uroman.pl