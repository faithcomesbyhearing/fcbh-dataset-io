import os
import sys
import torch
from torch.utils.data import Dataset, DataLoader
from sqlite_utility import *
from io import BytesIO

class MyDataset(Dataset):
    def __init__(self, database):
        super().__init__()
        self.database = database

    def __len__(self):
        count = self.database.selectOne('SELECT count(*) FROM tensors', ())
        return count[0]

    def __getitem__(self, idx):
        query = 'SELECT input_values, attention_mask, labels, memory_mb FROM tensors WHERE idx = ?'
        (inputValues, attentionMask, labels, memoryMB) = self.database.selectOne(query, (idx,))
        inputBuffer = BytesIO(inputValues)
        audioTensor = torch.load(inputBuffer)
        maskBuffer = BytesIO(attentionMask)
        maskTensor = torch.load(maskBuffer)
        labelsBuffer = BytesIO(labels)
        labelsTensor = torch.load(labelsBuffer)
        print("tensor", idx, audioTensor.shape, type(audioTensor), audioTensor.nbytes / (1024**2))
        #print(reference, memoryMB, "MB")
        return {
            "input_values": audioTensor,
            "attention_mask": maskTensor,
            "labels": labelsTensor
        }

if __name__ == "__main__":
    dbPath = os.getenv("FCBH_DATASET_TMP") + "/N2KEUWB4.db"
    database = SqliteUtility(dbPath)
    dataset = MyDataset(database)
    length = len(dataset)
    print("length", length)
    for i in range(length):
        data = dataset[i]
        audioTensor = data["input_values"]
        print("audio", audioTensor.shape, type(audioTensor))
        labelsTensor = data["labels"]
        print("labels", labelsTensor.shape, type(labelsTensor), labelsTensor)
        #print("text", data["text"])
        #print("reference", data["reference"])
        #print("memory_mb", data["memory_mb"])
    database.close()