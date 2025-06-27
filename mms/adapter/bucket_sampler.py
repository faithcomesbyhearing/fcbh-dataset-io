import torch
from torch.utils.data import Sampler
import numpy as np

class BucketSampler(Sampler):
    def __init__(self, dataset, target_memory_mb=1000, max_batch_size=32):
        self.dataset = dataset
        self.target_memory_mb = target_memory_mb
        self.max_batch_size = max_batch_size

        # Calculate sequence lengths
        self.lengths = []
        for i in range(len(dataset)):
            item = dataset[i]
            memoryEst = item['num_bytes'] * 4 / (1024 * 1024)
            self.lengths.append((i, memoryEst))

    def __iter__(self):
        batches = []
        current_batch = []
        current_memory = 0

        for idx, memoryEst in self.lengths:
            # estimated_memory = length * 4 / (1024 * 1024)  # 4 bytes per float, convert to MB

            # Check if adding this sample exceeds limits
            if (current_memory + memoryEst > self.target_memory_mb or
                len(current_batch) >= self.max_batch_size) and current_batch:
                batches.append(current_batch)
                current_batch = []
                current_memory = 0

            current_batch.append(idx)
            current_memory += memoryEst

        # Don't forget the last batch
        if current_batch:
            batches.append(current_batch)

        # Shuffle batches
        np.random.shuffle(batches)

        for batch in batches:
            yield batch

    def __len__(self):
        return len(self.dataset)