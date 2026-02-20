# Installation Instructions for Artificial Polyglot (Arti)

These instructions have been used with MacOS, Centos Linux, and Ubuntu Linux.
Installing on Windows might be easiest with Docker.

## Download and install go

For instruction on the go language, see <https://go.dev/learn>

To install go, download the binary at <https://go.dev/doc/install>, and follow the installation instructions.

Open a new terminal window to verify that go is installed.
```
go version
which go
```
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
conda init
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

** This is not working, and not needed.  So, skip it.
```
cd ~/go/fcbh-dataset-io
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
cd ~/go/fcbh-dataset-io
sh mms/adapter/build_env.sh
```

### Install packages for the MMS ASR component
```
cd ~/go/fcbh-dataset-io
sh mms/mms_asr/build_env.sh
```

### Install packages from the MMS Forced Alignment component
```
cd ~/go/fcbh-dataset-io
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

## Compile Arti Server, and start the server

Arti is run in production as a server, that listens for request files to be added to an S3 bucket
that functions as a queue.

```
cd ~/go/fcbh-dataset-io
go install ./controller/queue_server
cd
nohup ~/go/bin/queue_server &
```

## Compile and run Arti as a command line program

Arti can also be run as a command line program.

```
cd ~/go/fcbh-dataset-io
go install ./controller/dataset_cli
$HOME/go/bin/dataset_cli request_config.yaml
```

## Development and testing

The following is a sample test.  This request.yaml file is stating that text data is to be loaded from
an Excel spreadsheet located on AWS S3.  The audio files in .wav format are also located on AWS S3.
The process first runs forced alignment to get timestamp and probability of correctness scores.
Then, the text and data is used to train an mms adapter model in the language.  Following that,
speech to text is run to create a transcript of the audio using the trained model.  Finally, the
transcript is compared to the original text, and a report is produced for analysts to review.  In this
case only one person is receiving email notification.

The SqliteTest struct provides a means to add database queries and expected results that will be 
run at completion of the test.

```
import (
	"testing"
)

const runAnything = `is_new: yes
dataset_name: ART
language_iso: spa
username: GaryNTest
notify_ok: [gary@shortsands.com]
notify_err: [gary@shortsands.com]
text_data:
  aws_s3: s3://dataset-vessel/vessel/ART_12231842/ART Text/XLSX/Arti Test_ART_LineBased.xlsx
audio_data:
  aws_s3: s3://dataset-vessel/vessel/ART_12231842/ART Line VOX/*.wav
timestamps:
  mms_align: y
training:
  redo_training: yes
  mms_adapter:
    batch_mb: 4
    num_epochs: 16
    learning_rate: 1e-3
    warmup_pct: 12.0
    grad_norm_max: 0.4
speech_to_text:
  adapter_asr: yes
compare:
  html_report: yes
  gordon_filter: 0
  compare_settings:
    lower_case: yes
    remove_prompt_chars: yes
    remove_punctuation: yes
    double_quotes:
      remove: yes
    apostrophe:
      remove: yes
    hyphen:
      remove: yes
    diacritical_marks:
      normalize_nfc: yes
`

func TestRunAnything(t *testing.T) {
	DirectSqlTest(runAnything, []SqliteTest{}, t)
}
```












