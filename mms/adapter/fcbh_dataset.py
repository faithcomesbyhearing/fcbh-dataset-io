import os
import torch
import torchaudio
from torch.utils.data import Dataset, DataLoader
import soundfile
from transformers import Wav2Vec2ForCTC, Wav2Vec2Config, Wav2Vec2Processor
import numpy as np
from sqlite_utility import *


class FCBHDataset(Dataset):
    def __init__(self, databasePath, audioDir, wav2Vec2Processor):
        self.databasePath = databasePath
        self.audioDir = audioDir
        self.wav2Vec2Processor = wav2Vec2Processor
        self.database = SqliteUtility(databasePath)
        maxDuration = self.database.selectOne("SELECT MAX(script_end_ts-script_begin_ts) FROM scripts", ())[0]
        self.maxLen = int(maxDuration * 16000)
        query = """SELECT audio_file, script_text, script_begin_ts, script_end_ts
                FROM scripts WHERE verse_str != '0'
                AND fa_score > 0.4
                AND script_id NOT IN
                (SELECT distinct script_id FROM words WHERE ttype='W' AND fa_score < 0.01)"""
        self.data = self.database.select(query,())
        # Get vocabularySize
        chars = set()
        chars.add(' ')
        words = self.database.select("SELECT word FROM words WHERE ttype='W'", ())
        for wd in words:
            for ch in wd[0].lower():
                chars.add(ch)
        chars.discard('\u2014') # another hyphen
        chars.discard('\u002d') # hyphen
        self.vocabularySize = len(chars)
        self.database.close()


    def getVocabularySize(self):
        return self.vocabularySize


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

        # Pad or trim audio
        if len(speech) < self.maxLen:
            padded_speech = np.zeros(self.maxLen)
            padded_speech[:len(speech)] = speech
            speech = padded_speech
        else:
            speech = speech[:self.maxLen]

        # Prepare audio
        audioTensor = self.wav2Vec2Processor(speech, sampling_rate=16000, return_tensors="pt").input_values.squeeze()

        # Prepare text
        processed = self.wav2Vec2Processor(text=text)
        labelTensor = torch.tensor(processed.input_ids).squeeze()

        return audioTensor, labelTensor, text


if __name__ == "__main__":
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2ENGWEB.db"
    audioPath = os.getenv("FCBH_DATASET_FILES") + "/ENGWEB/ENGWEBN2DA"
    model_name = "facebook/mms-1b-all"
    wav2Vec2Processor = Wav2Vec2Processor.from_pretrained(model_name)
    data = FCBHDataset(dbPath, audioPath, wav2Vec2Processor)
    length = data.__len__()
    print("length", length)
    (audioTensor, labelTensor, text) = data.__getitem__(0)
    print("audio:", audioTensor)
    print("labels:", labelTensor)
    print("text:", text)
    print("vocabSize", data.getVocabularySize())