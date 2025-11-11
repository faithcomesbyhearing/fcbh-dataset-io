import os
import sys
import torch
import numpy as np
from torch.utils.data import Dataset, DataLoader
from sqlite_utility import *
from data_preparation import decompress
from io import BytesIO

class MyDataset(Dataset):
    def __init__(self, database):
        super().__init__()
        self.database = database

    def __len__(self):
        count = self.database.selectOne('SELECT count(*) FROM tensors', ())
        return count[0]

    def __getitem__(self, idx):
        query = 'SELECT num_samples, input_values, attention_mask, labels, memory_mb FROM tensors WHERE idx = ?'
        (numSamples, inputValues, attentionMask, labels, memoryMB) = self.database.selectOne(query, (idx,))
        audioTensor = decompress(inputValues, np.float32, numSamples)
        maskTensor = decompress(attentionMask, np.int32, numSamples)
        labelsTensor = decompress(labels, np.int64, numSamples)
        print("tensor", idx, audioTensor.shape, type(audioTensor), audioTensor.nbytes / (1024**2))
        return {
            "input_values": audioTensor,
            "attention_mask": maskTensor,
            "labels": labelsTensor
        }

if __name__ == "__main__":
    dbPath = os.getenv("FCBH_DATASET_TMP") + "/N2KEUWB4_CTC.db"
    database = SqliteUtility(dbPath)
    dataset = MyDataset(database)
    print(f"Dataset size: {len(dataset)}")
    dataloader = DataLoader(dataset, batch_size=None, shuffle=True)
    for step, batch in enumerate(dataloader):
        print(f"\n=== Batch {step} ===")
        print(f"Batch type: {type(batch)}")
        print(f"Batch keys: {batch.keys() if isinstance(batch, dict) else 'N/A'}")

        # Get dimensions to understand batch structure
        print(f"input_values shape: {batch['input_values'].shape}")

        # Extract FIRST sample from batch
        values = batch['input_values'][10]
        mask = batch['attention_mask'][10]

        print(f"\n=== Single Sample ===")
        print(f"Sample input shape: {values.shape}")
        print(f"Sample mask shape: {mask.shape}")

        transition = (mask[:-1] != mask[1:]).nonzero()
        if len(transition) > 0:
            boundary = transition[0].item()
            start = max(0, boundary - 10)
            end = min(len(values), boundary + 10)

            print(f"\n=== Boundary at index {boundary} ===")
            for i in range(start, end):
                marker = " <-- BOUNDARY" if i == boundary else ""
                print(f"Index {i:4d}: value={values[i]:8.4f}, mask={mask[i]}{marker}")
        else:
            print("No boundary found - mask is all 1s or all 0s")

        # Stop after first few batches
        if step >= 2:
            break
    database.close()