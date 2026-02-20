#!/bin/bash

# https://github.com/readbeyond/aeneas

conda create -y -n aeneas python=3.8

conda activate aeneas

#conda install -y ffmpeg -c conda-forge  ## appears to be redundant

conda install -y numpy -c conda-forge

# debian
#sudo apt-get -y install espeak libespeak-dev
# centos
# sudo yum -y install espeak espeak-devel

conda install -y "setuptools <60"

pip install aeneas

python -m aeneas.diagnostics

conda deactivate