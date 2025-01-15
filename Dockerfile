FROM public.ecr.aws/docker/library/golang:1.23.4-alpine as builder

WORKDIR /app

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go application
RUN go build -o taskService controller/ecs_entry.go

# Deployment stage
FROM public.ecr.aws/amazonlinux/amazonlinux:latest

WORKDIR /root/


# Install necessary tools
RUN yum update -y && \
    # yum install -y wget tar xz && \
    yum clean all


# TODO: include build instructions from docs/user_data.sh

# Copy the built binary
COPY --from=builder /app/taskService .

# Ensure the binary is executable
RUN chmod +x ./taskService

# Use environment variable to determine the service to run
ENTRYPOINT ["./taskService"]

# Dataset_server build
FROM ubuntu:20.04

RUN apt-get update
RUN apt-get -y upgrade

# create user
RUN adduser dataset
RUN usermod -aG sudo dataset 
RUN su - dataset

# install dependencies
RUN apt-get install -y wget
RUN apt-get install -y git

# install aeneas dependencies
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y ffmpeg
RUN apt-get install -y espeak espeak-data libespeak1 libespeak-dev
RUN apt-get install -y festival*
RUN apt-get install -y build-essential
RUN apt-get install -y flac libasound2-dev libsndfile1-dev vorbis-tools
RUN apt-get install -y libxml2-dev libxslt-dev zlib1g-dev
RUN apt-get install -y python-dev python3-pip
RUN pip3 install numpy

RUN pip3 install aeneas
RUN pip3 install librosa
RUN pip3 install -U openai-whisper

RUN pip3 install fasttext
RUN apt-get install sqlite3

# install go
RUN wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

# install application server
RUN git clone https://github.com/garygriswold/fcbh-dataset-io.git
WORKDIR /fcbh-dataset-io

# setup env vars
ENV FCBH_DBP_KEY=b4715786-9b8e-4fbe-a9b9-ff448449b81b
ENV FCBH_DATASET_DB=/home/dataset/data
ENV FCBH_DATASET_FILES=/home/dataset/data/download
ENV FCBH_DATASET_TMP=/home/dataset/data/tmp
ENV PYTHON_EXE=/usr/bin/python3
ENV WHISPER_EXE=/home/dataset/.local/bin/whisper
ENV FCBH_DATASET_LOG_FILE=/home/dataset/dataset.log
ENV FCBH_DATASET_LOG_LEVEL=DEBUG

ENV GOBIN=/
RUN go install controller/api_server/dataset_server.go
CMD /dataset_server
