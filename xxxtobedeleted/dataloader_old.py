import os
import torch
from torch.utils.data import DataLoader, Subset


def MyDataLoader(dataset, loadType, batchSize):
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

    if batchSize == 1:
        loader = DataLoader(
            dataset,
            batch_size=1,
            shuffle=shuffle,
            num_workers=0
        )
    else:
        loader = DataLoader(
            dataset,
            batch_size=batchSize,
            shuffle=shuffle,
            num_workers=0,
            collate_fn=collate_batch
        )
    return loader


## This has NOT been tested
def collate_batch(batch):
    inputValues, labels, texts = zip(*batch)

    # Get the original lengths before padding
    inputLengths = [len(x) for x in inputValues]
    maxInputLen = max(inputLengths)

    # Now create attention masks with the correct size
    attentionMasks = []
    for length in inputLengths:
        mask = torch.zeros(maxInputLen, dtype=torch.long)
        mask[:length] = 1
        attentionMasks.append(mask)
    attentionMasksBatch = torch.stack(attentionMasks)

    # Pad input values
    inputValuesBatch = torch.nn.utils.rnn.pad_sequence(
        inputValues,
        batch_first=True
    )
    inputValuesBatch = torch.stack(inputValues)


    labelsBatch = torch.nn.utils.rnn.pad_sequence(
        labels,
        batch_first=True
    )
    return inputValuesBatch, attentionMasksBatch, labelsBatch, texts


if __name__ == "__main__":
    from transformers import Wav2Vec2ForCTC, Wav2Vec2Config, Wav2Vec2Processor
    from dataset import *
    from sqlite_utility import *
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2ENGWEB.db"
    audioPath = os.getenv("FCBH_DATASET_FILES") + "/ENGWEB/ENGWEBN2DA"
    model_name = "facebook/mms-1b-all"
    wav2Vec2Processor = Wav2Vec2Processor.from_pretrained(model_name)
    database = SqliteUtility(dbPath)
    dataPruner(database)
    dataset = MyDataset(database, audioPath, wav2Vec2Processor)
    dataLoader = MyDataLoader(dataset, "full", 1)
    print(dataLoader, type(dataLoader))
    #print(dir(dataLoader))
