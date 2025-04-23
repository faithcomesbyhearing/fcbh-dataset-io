#!/bin/bash

# This is meant to be executed line by line

conda deactivate
conda remove --name mms_adapter --all

conda create --name mms_adapter python=3.11 -y
conda activate mms_adapter
pip install --upgrade pip
pip install numpy
pip install torch
pip install torchaudio
pip install soundfile
pip install adapters
pip install peft
pip install jiwer
pip install unicodedata
#pip install accelerate


# Ran on Mac 4/9/25