#!/bin/bash

# https://pytorch.org/audio/main/tutorials/forced_alignment_for_multilingual_data_tutorial.html

# PREREQUISITE: Install ffmpeg and sox
# CentOS: sudo yum install ffmpeg sox

# PREREQUISITE: The MMS model will be downloaded automatically on first use
# If download fails, manually download with:
# python3 -c "from torchaudio.pipelines import MMS_FA as bundle; bundle.get_model(with_star=False)"

conda create -y -n mms_fa python=3.11

conda activate mms_fa

conda install -y pytorch torchaudio pytorch-cuda=12.1 -c pytorch -c nvidia
# On Mac
# conda install -y pytorch::pytorch torchaudio -c pytorch

conda install -y pysoundfile -c conda-forge

conda install -y ffmpeg-python -c conda-forge

conda install -y sox -c conda-forge
pip install sox

pip install uroman
cp /opt/conda/envs/mms_fa/bin/uroman /opt/conda/envs/mms_fa/bin/uroman.pl
# on Mac
# cp /Users/gary/miniforge3/envs/mms_fa/bin/uroman /Users/gary/miniforge3/envs/mms_fa/bin/uroman.pl

conda deactivate
