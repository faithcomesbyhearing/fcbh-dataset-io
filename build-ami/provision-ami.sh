#!/bin/bash -x
 
# ---------- add ec2-user ----------
# useradd -m -s /bin/bash ec2-user
 
# mkdir -p /home/ec2-user/.ssh
# chmod 700 /home/ec2-user/.ssh
 
# cp /home/ubuntu/.ssh/authorized_keys /home/ec2-user/.ssh/
# chown -R ec2-user:ec2-user /home/ec2-user/.ssh
# chmod 600 /home/ec2-user/.ssh/authorized_keys
 
# # Assign sudo privilege
# usermod -aG sudo ec2-user
 
# # Make ec2-user passwordless sudo (like Amazon Linux)
# echo "ec2-user ALL=(ALL) NOPASSWD:ALL" | sudo tee /etc/sudoers.d/90-ec2-user
# chmod 440 /etc/sudoers.d/90-ec2-user
 
# chown -R ec2-user:ec2-user /opt/conda
# chmod -R u+rwX,go-w /opt/conda
 
# # Switch to ec2-user for the remaining operations
# sudo -u ec2-user bash << 'EOF'
 
set -x
 
cd $HOME
 
# ==> WARNING: A newer version of conda exists. <==
#     current version: 24.11.2
#     latest version: 25.1.0

# Please update conda by running

#     $ conda update -n base -c conda-forge conda

/opt/conda/bin/conda init bash
source /opt/conda/bin/activate
 
# ---------- base ----------
 
conda activate base
 
# Install executables using conda into the base environment
conda install -y ffmpeg -c conda-forge
conda install -y sox -c conda-forge
conda install -y sqlite -c conda-forge
 
sudo apt-get -y update
sudo apt-get -y upgrade
 
# ---------- aeneas ----------
 
# https://github.com/readbeyond/aeneas
 
conda create -y -n aeneas python=3.8
# Daemons using outdated libraries
 
conda activate aeneas
 
conda install -y ffmpeg -c conda-forge  ## appears to be redundant
 
conda install -y numpy -c conda-forge
 
sudo apt-get -y install espeak libespeak-dev mysql-client
# centos
# sudo yum -y install espeak espeak-devel
 
conda install -y "setuptools <60"
# results
# LibMambaUnsatisfiableError: Encountered problems while solving:
#   - package setuptools-23.0.0-py27_0 requires python 2.7*, but none of the providers can be installed

# Could not solve for environment specs
# The following packages are incompatible
# ├─ pin-1 is installable and it requires
# │  └─ python 3.12.* , which can be installed;
# └─ setuptools <60  is not installable because there are no viable options
#    ├─ setuptools [23.0.0|25.1.3|...|44.0.0] would require
#    │  └─ python [2.7* |>=2.7,<2.8.0a0 ], which conflicts with any installable versions previously reported;
#    ├─ setuptools [23.0.0|25.1.3|...|32.3.1] would require
#    │  └─ python 3.4* , which conflicts with any installable versions previously reported;
#    ├─ setuptools [23.0.0|25.1.3|...|40.4.3] would require
#    │  └─ python [3.5* |>=3.5,<3.6.0a0 ], which conflicts with any installable versions previously reported;
#    ├─ setuptools [32.3.0|32.3.1|...|39.0.1] would require
#    │  └─ python 3.6* , which conflicts with any installable versions previously reported;
#    ├─ setuptools [39.1.0|39.2.0|...|58.0.4] would require
#    │  └─ python >=3.6,<3.7.0a0 , which conflicts with any installable versions previously reported;
#    ├─ setuptools [40.0.0|40.2.0|...|59.8.0] would require
#    │  └─ python >=3.7,<3.8.0a0 , which conflicts with any installable versions previously reported;
#    ├─ setuptools [41.6.0|42.0.0|...|59.8.0] would require
#    │  └─ python >=3.8,<3.9.0a0 , which conflicts with any installable versions previously reported;
#    ├─ setuptools [49.6.0|57.4.0|...|59.8.0] would require
#    │  └─ python >=3.9,<3.10.0a0 , which conflicts with any installable versions previously reported;
#    ├─ setuptools [57.4.0|58.4.0|...|59.8.0] would require
#    │  └─ python >=3.10,<3.11.0a0 , which conflicts with any installable versions previously reported;
#    └─ setuptools [57.4.0|59.8.0] would require
#       └─ python >=3.11,<3.12.0a0 , which conflicts with any installable versions previously reported.


pip install aeneas
#results
#   Downloading aeneas-1.7.3.0.tar.gz (5.5 MB)
#      ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 5.5/5.5 MB 97.6 MB/s eta 0:00:00
#   Preparing metadata (setup.py) ... error
#   error: subprocess-exited-with-error
  
#   × python setup.py egg_info did not run successfully.
#   │ exit code: 1
#   ╰─> [3 lines of output]
#       [ERRO] You must install numpy before installing aeneas
#       [INFO] Try the following command:
#       [INFO] $ sudo pip install numpy
#       [end of output]
 
python -m aeneas.diagnostics
 
conda deactivate
 
# ---------- fasttext ----------
 
conda create -y -n fasttext --no-default-packages
 
conda activate fasttext
 
cd /opt/conda/envs/fasttext
# on Mac
# cd $HOME/miniforge3/envs/fasttext/
 
git clone https://github.com/facebookresearch/fastText.git
cd fastText
make
 
conda deactivate
 
# ---------- librosa ----------
 
conda create -y -n librosa python=3.11
 
conda activate librosa
 
pip install librosa
 
conda deactivate
 
cd $HOME
 
# ---------- mms_asr ----------
 
conda create -y -n mms_asr python=3.11
 
conda activate mms_asr
 
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
 
conda deactivate
 
# ---------- mms_fa ----------
 
# https://pytorch.org/audio/main/tutorials/forced_alignment_for_multilingual_data_tutorial.html
 
conda create -y -n mms_fa python=3.11
 
conda activate mms_fa
 
conda install -y pytorch torchaudio pytorch-cuda=12.1 -c pytorch -c nvidia
# On Mac
# conda install -y pytorch::pytorch torchaudio -c pytorch
 
conda install -y pysoundfile -c conda-forge
 
conda install -y ffmpeg-python -c conda-forge
 
pip install sox
 
pip install uroman
cp /opt/conda/envs/mms_fa/bin/uroman /opt/conda/envs/mms_fa/bin/uroman.pl
# on Mac
# cp /Users/gary/miniforge3/envs/mms_fa/bin/uroman /Users/gary/miniforge3/envs/mms_fa/bin/uroman.pl
 
conda deactivate
 
# ---------- whisper ----------
 
# https://pypi.org/project/openai-whisper/
 
conda create -y -n whisper python=3.11
 
conda activate whisper
 
pip install -U openai-whisper
 
conda deactivate
 
# ---------- go ----------
#sudo yum -y install golang
# wget https://go.dev/dl/go1.23.3.linux-amd64.tar.gz
# sudo tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz
# export PATH=/usr/local/go/bin:$PATH
# rm go1.23.3.linux-amd64.tar.gz
sudo snap install go --classic

# Check go
# which go
# go version
# go env
 
# # Set GOPATH
# cd $HOME
# mkdir go
# export GOPATH=$HOME/go
 
# Cloudwatch agent 
# the agent is pre-installed on ubuntu 22.04 from AWS
sudo systemctl enable amazon-cloudwatch-agent

# SSM agent
# the agent is pre-installed on ubuntu 22.04 via snap(need to verify)
# sudo systemctl status snap.amazon-ssm-agent.amazon-ssm-agent.service

# Shell connection
# ssh -i ~/.ssh/deep-learning-dev.pem ubuntu@<IP address>
# or
# Via AWS Console, specify EC2 instance, then select EC2 Instance Connect with default user "root"
# after connecting, "su ubuntu"