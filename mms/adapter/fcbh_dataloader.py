import os
import torch
import soundfile
from torch.utils.data import DataLoader, Subset


def FCBHDataLoader(dataset, loadType, batchSize, numWorkers):
    datasetSize = len(dataset)
    # Create indices lists
    if loadType == 'train':
        indices = [i for i in range(datasetSize) if i % 5 != 0]
        dataset = Subset(dataset, indices)
        shuffle = True
    elif loadType == 'test':
        indices = [i for i in range(datasetSize) if i % 5 == 0]
        dataset = Subset(dataset, indices)
        shuffle = False
    else: # loadType == 'full'
        dataset = dataset
        shuffle = True

    loader = DataLoader(
        dataset,
        batch_size=batchSize,
        shuffle=shuffle,
        num_workers=numWorkers,
        collate_fn=collate_batch
    )
    return loader


def collate_batch(batch):
    # Separate audio and label tensors
    audio_tensors, label_tensors, texts = zip(*batch)

    # Stack them into batches
    audio_batch = torch.stack(audio_tensors)
    label_batch = torch.stack(label_tensors)
    return audio_batch, label_batch, texts


if __name__ == "__main__":
    from transformers import Wav2Vec2ForCTC, Wav2Vec2Config, Wav2Vec2Processor
    from fcbh_dataset import *
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2ENGWEB.db"
    audioPath = os.getenv("FCBH_DATASET_FILES") + "/ENGWEB/ENGWEBN2DA"
    model_name = "facebook/mms-1b-all"
    wav2Vec2Processor = Wav2Vec2Processor.from_pretrained(model_name)
    data = FCBHDataset(dbPath, audioPath, wav2Vec2Processor)
    dataset = FCBHDataLoader(data, "full", 1, 1)
    print("dataset", dataset)