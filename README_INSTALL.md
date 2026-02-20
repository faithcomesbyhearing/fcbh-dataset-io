# Installation Instructions for Artificial Polyglot (Arti)

These instructions have been used with MacOS, Centos Linux, and Ubuntu Linux.
Installing on Windows might be easiest with Docker.

## Download and install go

For instruction on the go language, see <https://go.dev/learn>

To install go, download the binary at <https://go.dev/doc/install>, and follow the installation instructions.

## Download the Arti repository

cd to the location where you want to install Arti.  Optionally, create a parent directory for 
your go project workspace.  Here I install the project in a directory named 'go' in my home directory.

```
cd
mkdir go
cd go
git clone https://github.com/faithcomesbyhearing/fcbh-dataset-io
```

## Download and install Miniforge

Conda is used to solve two problems.  It provides python environment management, which is needed
because of incompatibilities between the python versions used by various Arti modules.
Second, it provides a package manager that will install identical packages for Linux, MacOS, and Windows.
So, that the installation instructions can be nearly identical for all operating systems.

Information about Miniforge can be found at <https://github.com/conda-forge/miniforge>.

Download and install Miniforge for your operating system: <https://conda-forge.org/download/>.

Once installed, type 'conda init' to initialize conda in your shell.

```
init conda
```
After doing this, your command line prompt should start with '(base)'.
The command syntax for using Miniforge is identical to conda.

## Create conda environments, and install packages

### Install commonly needed packages into the base environment.
```
conda activate base
conda install -y ffmpeg -c conda-forge
conda install -y sox -c conda-forge
conda install -y sqlite -c conda-forge
```

### Install packages for the aeneas component 
```
cd fcbh-dataset-io
sh encode/build_aeneas_env.sh
```

### Install the fasttext executable
```
conda create -y -n fasttext python=3.11
cd ~/miniforge/envs/fasttext
git clone https://github.com/facebookresearch/fastText.git
cd fastText
make
```

### Install packages for the librosa component
```
conda create -y -n librosa python=3.11
conda activate librosa
pip install librosa
conda deactivate
```

### Install packages for the MMS Adapter Training component
```
cd fcbh-dataset-io
sh mms/adapter/build_env.sh
```

### Install packages for the MMS ASR component
```
cd fcbh-dataset-io
sh mms/mms_asr/build_env.sh
```

### Install packages from the MMS Forced Alignment component
```
cd fcbh-dataset-io
sh mms/mms_align/build_env.sh
```

### Install Whisper for the Whisper component
```
# https://pypi.org/project/openai-whisper/
conda create -y -n whisper python=3.11
conda activate whisper
pip install -U openai-whisper
conda deactivate
```

## Add the environment variables to the correct profile file.
Modern MacOS: ~/.zprofile
Older MacOS: ~/.bash_profile
Ubuntu: ~/.bash_profile

### Conda environment variables

If you followed the instructions for installing packages, then set these environment variables 
as shown.  Or, adjust to suit the actual location of your conda environments.

```
export PY_ENV=$HOME/miniforge/envs
export FCBH_AENEAS_PYTHON=$PY_ENV/aeneas/bin/python
export FCBH_FASTTEXT_EXE=$PY_ENV/fasttext/fastText/fasttext
export FCBH_LIBROSA_PYTHON=$PY_ENV/librosa/bin/python
export FCBH_MMS_ASR_PYTHON=$PY_ENV/mms_asr/bin/python
export FCBH_MMS_FA_PYTHON=$PY_ENV/mms_fa/bin/python
export FCBH_MMS_ADAPTER_PYTHON=$PY_ENV/mms_adapter/bin/python
export FCBH_UROMAN_EXE=$PY_ENV/mms_fa/bin/uroman.pl
export FCBH_WHISPER_EXE=$PY_ENV/whisper/bin/whisper
```

### Directory path and file path variables
```
export GOPATH= # set to the directory where the Arti project is installed
export GOPROJ= # set to the location of the fcbh-dataset-io directory
export FCBH_DATASET_DB= # directory where each run's database will be stored.
export FCBH_DATASET_FILES= # directory where downloaded files will be stored.
export FCBH_DATASET_TMP= # directory where tmp files and tmp directories will be created.
export FCBH_DATASET_LOG_FILE= # path of the database log
export FCBH_DATASET_LOG_LEVEL= # set to DEBUG, INFO, or WARN to control the amount of logging.
```

### AWS S3 bucket names
```
export FCBH_DATASET_QUEUE= # When the server is to be started by a queue, this is the S3 bucket which has the queue.
export FCBH_DATASET_IO_BUCKET= # The name of the S3 bucket to receive each run's output.
```

### Email Notification

In production runs and some test runs, the server sends email notification when a job is complete,
with output or error logs attached.

```
export SMTP_SENDER_EMAIL= # The email username
export SMTP_PASSWORD= # The email password
export SMTP_HOST_NAME= # The SMTP server hostname
export SMTP_HOST_PORT= # The SMTP port
```

### Keys 

```
export FCBH_DBP_KEY= # This must be set with a Bible Brain key if you want to access FCBH Bible Brain.
```

This looks like an error.
export PATH=$PATH:/Users/gary/Library/Python/3.9/bin:$GOPROJ/bin











