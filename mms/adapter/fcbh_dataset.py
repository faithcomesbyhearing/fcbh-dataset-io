import os
import torch
import torchaudio
from torch.utils.data import Dataset, DataLoader
import soundfile
from transformers import Wav2Vec2ForCTC, Wav2Vec2Config, Wav2Vec2Processor
import numpy as np
from sqlite_utility import *


class FCBHDataset(Dataset):
    def __init__(self, databasePath, audioDir, processor):
        super(FCBHDataset).__init__()
        self.databasePath = databasePath
        self.audioDir = audioDir
        self.processor = processor
        self.database = SqliteUtility(databasePath)
        maxDuration = self.database.selectOne("SELECT MAX(script_end_ts-script_begin_ts) FROM scripts", ())[0]
        self.maxLen = int(maxDuration * 16000)
        query = """SELECT audio_file, script_text, script_begin_ts, script_end_ts
                FROM scripts WHERE verse_str != '0'
                AND fa_score > 0.4
                AND script_id NOT IN
                (SELECT distinct script_id FROM words WHERE ttype='W' AND fa_score < 0.01)"""
        self.data = self.database.select(query,())
        self.database.close()


    def __len__(self):
        return len(self.data)


    def __getitem__(self, idx):
        (audioFile, text, beginTS, endTS) = self.data[idx]
        audioFile = audioFile.replace(".mp3", ".wav")
        audioPath = os.path.join(self.audioDir, audioFile)

        # Load audio portion for script line
        info = torchaudio.info(audioPath, format="wav")
        speech, sample_rate = torchaudio.load(
            audioPath,
            frame_offset = int(beginTS * info.sample_rate),
            num_frames = int((endTS - beginTS) * info.sample_rate)
        )
        speech = speech.squeeze().numpy()
        # Normalize
        #speech = speech / (np.max(np.abs(speech)) + 1e-5)  # Normalize to [-1, 1]

         # Store original length for creating attention mask
        originalLength = len(speech)

        # Pad or trim audio
        if originalLength < self.maxLen:
            paddedSpeech = np.zeros(self.maxLen)
            paddedSpeech[:len(speech)] = speech
            speech = paddedSpeech
        else:
            speech = speech[:self.maxLen]

        attentionMask = np.ones(self.maxLen)
        if originalLength < self.maxLen:
            attentionMask[originalLength:] = 0

        # Prepare audio
        inputValues = self.processor(
                speech,
                sampling_rate=16000,
                return_tensors=None,
                padding=False
            ).input_values
        print("input", type(inputValues))

        # Prepare text
        #with self.processor.as_target_processor():
        labels = self.processor(text=text).input_ids

        # Now convert to tensors - this is fine for fixed-length data
        inputValuesTensor = torch.tensor(inputValues, dtype=torch.float).squeeze(0)
        attentionMaskTensor = torch.tensor(attentionMask, dtype=torch.long)
        labelsTensor = torch.tensor(labels, dtype=torch.long)

        return inputValuesTensor, attentionMaskTensor, labelsTensor, text


if __name__ == "__main__":
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2ENGWEB.db"
    audioPath = os.getenv("FCBH_DATASET_FILES") + "/ENGWEB/ENGWEBN2DA"
    model_name = "facebook/mms-1b-all"
    wav2Vec2Processor = Wav2Vec2Processor.from_pretrained(model_name)
    data = FCBHDataset(dbPath, audioPath, wav2Vec2Processor)
    length = data.__len__()
    print("length", length)
    (audioTensor, maskTensor, labelsTensor, text) = data.__getitem__(0)
    print("audio", audioTensor, audioTensor.shape, type(audioTensor))
    print("mask:", maskTensor, maskTensor.shape, type(maskTensor))
    print("labels", labelsTensor, labelsTensor.shape, type(labelsTensor))
    print("text", text)