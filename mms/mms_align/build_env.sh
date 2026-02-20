#!/bin/bash

# https://pytorch.org/audio/main/tutorials/forced_alignment_for_multilingual_data_tutorial.html

# PREREQUISITE: The MMS model will be downloaded automatically on first use
# If download fails, manually download with:
# python3 -c "from torchaudio.pipelines import MMS_FA as bundle; bundle.get_model(with_star=False)"

conda create -y -n mms_fa python=3.11

conda activate mms_fa

if [ "$(uname)" == "Darwin" ]; then
  pip install torch torchaudio
else
  pip install torch torchaudio --index-url https://download.pytorch.org/whl/cpu
fi
# conda install -y pytorch torchaudio pytorch-cuda=12.1 -c pytorch -c nvidia
# On Mac
# conda install -y pytorch::pytorch torchaudio -c pytorch

conda install -y pysoundfile -c conda-forge

conda install -y ffmpeg-python -c conda-forge

conda install -y sox -c conda-forge
pip install sox

pip install uroman
if [ "$(uname)" == "Darwin" ]; then
  cp $HOME/miniforge/envs/mms_fa/bin/uroman $HOME/miniforge/envs/mms_fa/bin/uroman.pl
else
  cp /opt/conda/envs/mms_fa/bin/uroman /opt/conda/envs/mms_fa/bin/uroman.pl
fi

conda deactivate
