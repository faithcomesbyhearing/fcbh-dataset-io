#!/bin/bash

# This is meant to be executed line by line

conda deactivate
conda remove --name mms_zero_shot --all

conda create --name mms_zero_shot python=3.11 -y
conda activate mms_zero_shot

pip install flashlight-text
pip install git+https://github.com/kpu/kenlm.git
pip install cmake
