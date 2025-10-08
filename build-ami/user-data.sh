#!/bin/bash -xv

mkdir -p $HOME/go/src
cd $HOME/go/src
rm -rf *
git clone https://github.com/faithcomesbyhearing/fcbh-dataset-io.git dataset
cd $HOME/go/src/dataset
go install dataset/controller/queue_server
cd $HOME
$HOME/go/bin/queue_server &