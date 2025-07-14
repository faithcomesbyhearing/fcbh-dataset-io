#!/bin/bash -v

conda activate mms_adapter

iso639=keu
database=$FCBH_DATASET_DB/GaryNTest/N2KEUWB4.db
audio_dir=$FCBH_DATASET_FILES/N2KEUWB4/N2KEUWBT
batch_size=4
num_epochs=1

time python trainer.py $iso639 $database $audio_dir $batch_size $num_epochs

#for file in *.mp3; do
#    ffmpeg -i "$file" -acodec pcm_s16le -ar 16000 -ac 1 "${file%.mp3}.wav"
#done