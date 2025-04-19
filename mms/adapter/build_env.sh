#!/bin/bash

# Source article is dated June 19, 2023

# This is meant to be executed line by line
# pip install transformers==4.26.0 datasets==2.6.1 torch==1.13.1 evaluate==0.4.0
conda deactivate
conda remove --name mms_adapter --all

conda create --name mms_adapter python=3.11 -y
conda activate mms_adapter
pip install --upgrade pip
pip install numpy
pip install torch
pip install torchaudio
pip install soundfile
#pip install transformers
#pip install adapter-transformers ## not sure this is needed
pip install adapters
pip install peft
pip install jiwer
#pip install datasets\[audio\]==2.6.1
#pip install evaluate==0.4.0
#pip install git+https://github.com/huggingface/transformers.git
#pip install transformers==4.26.0
#pip install jiwer
#pip install accelerate


# Ran on Mac 4/9/25