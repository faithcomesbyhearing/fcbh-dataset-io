#!/bin/bash

# This is meant to be executed line by line

conda deactivate
conda remove --name mms_zero_shot --all

conda create --name mms_zero_shot python=3.11 -y
conda activate mms_zero_shot

pip install flashlight-text
pip install git+https://github.com/kpu/kenlm.git
pip install cmake
pip install nltk

python -c "nltk.download('punkt')"
python -c "nltk.download('punkt_tab')"

conda install -y boost -c conda-forge
conda install -y cmake -c conda-forge

git clone https://github.com/kpu/kenlm
cd kenlm
mkdir -p build
cd build
cmake ..
make -j 4

## Needed for ASR

conda install -y pytorch torchaudio pytorch-cuda=12.1 -c pytorch -c nvidia
# On Mac
# conda install -y pytorch::pytorch torchaudio -c pytorch

pip install accelerate
pip install datasets
#pip install --upgrade transformers
pip install transformers
pip install soundfile
pip install librosa

pip install uroman
cp /opt/conda/envs/mms_asr/bin/uroman /opt/conda/envs/mms_asr/bin/uroman.pl
# on Mac
# cp /Users/gary/miniforge3/envs/mms_asr/bin/uroman /Users/gary/miniforge3/envs/mms_asr/bin/uroman.pl


