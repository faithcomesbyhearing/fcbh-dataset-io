#!/bin/bash

# Source article is dated June 19, 2023

# This is meant to be executed line by line
# pip install transformers==4.26.0 datasets==2.6.1 torch==1.13.1 evaluate==0.4.0
conda deactivate
conda remove --name mms_adapter --all

conda create --name mms_adapter python=3.9 -y
conda activate mms_adapter
pip install --upgrade pip
pip install datasets\[audio\]==2.6.1
pip install evaluate==0.4.0
#pip install git+https://github.com/huggingface/transformers.git
pip install transformers==4.26.0
pip install jiwer
pip install accelerate
#pip install torch==1.13.1



# Ran on Mac 4/9/25