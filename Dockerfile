FROM public.ecr.aws/ubuntu/ubuntu:20.04

WORKDIR /app

# Install base utilities
RUN apt-get update \
    && apt-get -y upgrade \
    && apt-get install -y wget \
    && apt-get install -y git \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y cmake \
    && apt-get install -y build-essential


# Install conda
ENV CONDA_DIR /opt/conda
RUN wget --quiet https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh -O ~/miniconda.sh && \
    /bin/bash ~/miniconda.sh -b -p /opt/conda

# Put conda in path 
ENV PATH=$CONDA_DIR/bin:$PATH

# ---------- base ----------
COPY docker_conda_env/environment_base.yml .
RUN conda env update -f environment_base.yml


# ---------- aeneas ----------
COPY docker_conda_env/environment_aeneas.yml .
RUN conda env create -f environment_aeneas.yml \
    && conda run -n aeneas apt-get install -y espeak libespeak-dev build-essential \
    && conda run -n aeneas pip install numpy \
    && conda run -n aeneas pip install aeneas


# ---------- fasttext ----------
RUN conda create -y -n fasttext --no-default-packages
RUN git clone https://github.com/facebookresearch/fastText.git
RUN mkdir build
WORKDIR /app/fastText/build
RUN cmake .. && make && make install 
WORKDIR /app


# ---------- librosa ----------
COPY docker_conda_env/environment_librosa.yml .
RUN conda env create -f environment_librosa.yml \
    && conda run -n librosa pip install librosa


# ---------- mms_asr ----------
COPY docker_conda_env/environment_mms_asr.yml .
RUN conda env create -f environment_mms_asr.yml \
    && conda run -n mms_asr pip install accelerate datasets soundfile librosa  \
    && conda run -n mms_asr pip install --upgrade transforms \
    && conda run -n mms_asr pip install uroman 
RUN cp /opt/conda/envs/mms_asr/bin/uroman /opt/conda/envs/mms_asr/bin/uroman.pl


# ---------- mms_fa ----------
COPY docker_conda_env/environment_mms_fa.yml .
RUN conda env create -f environment_mms_fa.yml \
    && conda run -n mms_fa pip install sox uroman
RUN cp /opt/conda/envs/mms_fa/bin/uroman /opt/conda/envs/mms_fa/bin/uroman.pl


# ---------- whisper ----------
COPY docker_conda_env/environment_whisper.yml .
RUN conda env create -f environment_whisper.yml \
    && conda run -n whisper pip install -U openai-whisper


# ---------- go ----------
RUN wget https://go.dev/dl/go1.23.3.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz
ENV PATH=/usr/local/go/bin:$PATH
RUN rm go1.23.3.linux-amd64.tar.gz && mkdir /app/go
ENV GOPATH=/app/go/


# ---------- application server ----------
# Copy the entire project
RUN mkdir /app/go/src
WORKDIR /app/go/src
COPY . .

WORKDIR /app
RUN mkdir /app/data && mkdir /app/data/download && mkdir /app/data/tmp


# ---------- environment variables ----------
ENV HOME=/app
ENV GOPROJ=$GOPATH/src
ENV PATH=$PATH:/usr/local/go/bin
ENV FCBH_DATASET_DB=$HOME/data
ENV FCBH_DATASET_FILES=$HOME/data/download
ENV FCBH_DATASET_TMP=$HOME/data/tmp
ENV FCBH_AENEAS_PYTHON=/opt/conda/envs/aeneas/bin/python
ENV FCBH_FASTTEXT_EXE=/opt/conda/envs/fasttext/fastText/fasttext
ENV FCBH_LIBROSA_PYTHON=/opt/conda/envs/librosa/bin/python
ENV FCBH_MMS_ASR_PYTHON=/opt/conda/envs/mms_asr/bin/python
ENV FCBH_MMS_FA_PYTHON=/opt/conda/envs/mms_fa/bin/python
ENV FCBH_UROMAN_EXE=/opt/conda/envs/mms_fa/bin/uroman
ENV FCBH_WHISPER_EXE=/opt/conda/envs/whisper/bin/whisper
ENV FCBH_DATASET_LOG_FILE=$HOME/dataset.log
ENV FCBH_DATASET_LOG_LEVEL=DEBUG
ENV PYTORCH_CUDA_ALLOC_CONF=expandable_segments:True



# Build the Go application
WORKDIR /app/go/src
RUN go install /app/go/src/controller/queue_server/s3_queue_server.go

# Ensure the binary is executable
RUN chmod +x $GOPATH/bin/s3_queue_server

# Use environment variable to determine the service to run
ENTRYPOINT $GOPATH/bin/s3_queue_server
