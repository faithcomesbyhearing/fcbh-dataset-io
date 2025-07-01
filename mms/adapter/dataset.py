import os
import sys
import torch
import torchaudio
from torch.utils.data import Dataset, DataLoader
import numpy as np
from sqlite_utility import *
from data_pruner import dataPruner


class MyDataset(Dataset):
    def __init__(self, database, audioDir, processor):
        super().__init__()
        self.database = database
        self.audioDir = audioDir
        self.processor = processor
        query = """
            SELECT s.book_id || ' ' || s.chapter_num || ':' || s.verse_str as ref,
                s.audio_file, s.script_begin_ts, s.script_end_ts, GROUP_CONCAT(w.word, ' ') AS text
            FROM scripts s
            JOIN words w ON w.script_id = s.script_id
            WHERE w.ttype = 'W' AND s.script_id IN (SELECT script_id FROM pruned_data)
            GROUP BY s.script_id, s.book_id, s.chapter_num, s.verse_str, s.audio_file, s.script_begin_ts, s.script_end_ts
            ORDER BY s.script_end_ts - s.script_begin_ts
            """
        self.data = self.database.select(query,())
        print("num lines", len(self.data))


    def __len__(self):
        return len(self.data)


    def __getitem__(self, idx):
        (reference, audioFile, beginTS, endTS, text) = self.data[idx]
        audioFile = audioFile.replace(".mp3", ".wav")
        audioPath = os.path.join(self.audioDir, audioFile)

        # Load audio portion for script line
        info = torchaudio.info(audioPath, format="wav")
        if info.sample_rate != 16000:
            print("Audio sample rate must be 16000", file=sys.stderr, flush=True)
            sys.exit(1)

        speech, sample_rate = torchaudio.load(
            audioPath,
            frame_offset = int(beginTS * 16000),
            num_frames = int((endTS - beginTS) * 16000)
        )
        speech = speech.squeeze().numpy()

        # Prepare audio
        inputValues = self.processor(
                speech,
                sampling_rate=16000,
                return_tensors=None,
                padding=False
            ).input_values
        inputValues = np.array(inputValues)
        inputValuesTensor = torch.tensor(inputValues, dtype=torch.float).squeeze(0)
        memoryBytes = inputValuesTensor.element_size() * inputValuesTensor.numel()
        if memoryBytes == 0:
            print(reference, audioFile, "Has Zero Length input_values", file=sys.stderr, flush=True)
            sys.exit(1)

        # Prepare text
        labels = self.processor(text=text).input_ids
        labelsTensor = torch.tensor(labels, dtype=torch.long)
        if memoryBytes == 0.0:
            print(reference, memoryBytes)

        return {
            "input_values": inputValuesTensor,
            "labels": labelsTensor,
            "text": text,
            "reference": reference,
            "memory_mb": memoryBytes / (1024 * 1024)
        }


if __name__ == "__main__":
    from tokenizer import createTokenizer
    from transformers import Wav2Vec2Processor, Wav2Vec2FeatureExtractor, Wav2Vec2CTCTokenizer
    from data_pruner import *
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2ENGWEB.db"
    database = SqliteUtility(dbPath)
    audioPath = os.getenv("FCBH_DATASET_FILES") + "/ENGWEB/ENGWEBN2DA"
    #model_name = "facebook/mms-1b-all"
    tokenizer = createTokenizer(database, "eng")
    featureExtractor = Wav2Vec2FeatureExtractor(
        feature_size=1, sampling_rate=16000, padding_value=0.0,
        do_normalize=True, return_attention_mask=True
    )
    processor = Wav2Vec2Processor(feature_extractor=featureExtractor, tokenizer=tokenizer)
    #wav2Vec2Processor = Wav2Vec2Processor.from_pretrained(model_name)
    dataPruner(database)
    dataset = MyDataset(database, audioPath, processor)
    database.close()
    length = len(dataset)
    print("length", length)
    for i in range(length):
        data = dataset[i]
        #data = dataset[1000]
        #(audioTensor, labelsTensor, text) = data.__getitem__(0)
        audioTensor = data["input_values"]
        print("audio", audioTensor.shape, type(audioTensor))
        labelsTensor = data["labels"]
        print("labels", labelsTensor.shape, type(labelsTensor), labelsTensor)
        print("text", data["text"])
        print("reference", data["reference"])
        print("memory_mb", data["memory_mb"])