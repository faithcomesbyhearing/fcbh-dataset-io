#!/bin/bash -v

conda activate mms_adapter

iso639=cul
#database=/Users/gary/FCBH2024/GaryNTest/N2CUL_MNT.db
database=/Users/gary/FCBH2024/GaryNTest/N2CUL_MNT_3vs.db
audio_dir=/Users/gary/FCBH2024/download/CULMNT/CULMNTN2DA
batch_size=2
num_epochs=1

python fcbh_train_adapter.py $iso639 $database $audio_dir $batch_size $num_epochs




#for file in *.mp3; do
#    ffmpeg -i "$file" -acodec pcm_s16le -ar 16000 -ac 1 "${file%.mp3}.wav"
#done