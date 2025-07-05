import torch
from torch.utils.data import Sampler
import numpy as np

class BucketSampler(Sampler):
    def __init__(self, dataset, target_memory_mb=1000, max_batch_size=32):
        self.dataset = dataset
        self.target_memory_mb = target_memory_mb
        self.max_batch_size = max_batch_size
        self.dataset_len = len(dataset)

    def __len__(self):
        return len(self.dataset)

    def __iter__(self):
        current_idx = 0

        while current_idx < self.dataset_len:
            batch = []
            current_memory = 0

            # Build one batch by calling dataset[i] only as needed
            while (current_idx < self.dataset_len and
                len(batch) < self.max_batch_size):

                # Only call dataset[i] when we need to check this sample
                sample = self.dataset[current_idx]
                memory_mb = sample['memory_mb']

                # Check if adding this sample would exceed memory limit
                if batch and (current_memory + memory_mb > self.target_memory_mb):
                    break # Don't include this sample, yield current batch
                # Add sample to current batch
                batch.append(current_idx)
                current_memory += memory_mb
                current_idx += 1

            if batch:
                yield batch
