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
        count = self.database.selectOne('SELECT count(*) FROM samples', ())
        return count[0]

    def __getitem__(self, idx):
        query = 'SELECT input_values, labels, text, reference, memory_mb FROM samples WHERE idx = ?'
        (inputValues, labels, text, reference, memoryMB) = self.database.selectOne(query, (idx,))
        inputBuffer = BytesIO(inputValues)
        audioTensor = torch.load(inputBuffer)
        labelsBuffer = BytesIO(labels)
        labelsTensor = torch.load(labelsBuffer)
        #print(reference, memoryMB, "MB")
        return {
            "input_values": audioTensor,
            "labels": labelsTensor,
            "text": text,
            "reference": reference,
            "memory_mb": memoryMB
        }

if __name__ == "__main__":
    dbPath = os.getenv("FCBH_DATASET_TMP") + "/N2ENGWEB.db"
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
        print("text", data["text"])
        print("reference", data["reference"])
        print("memory_mb", data["memory_mb"])
    database.close()