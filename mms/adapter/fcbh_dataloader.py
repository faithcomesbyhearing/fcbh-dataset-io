import os
from torch.utils.data import DataLoader, Subset


def FCBHDataLoader(dataset, kFold, batchSize, numWorkers):
    datasetSize = len(dataset)
    # Create indices lists
    testIndices = [i for i in range(datasetSize) if i % kFold == 0]
    trainIndices = [i for i in range(datasetSize) if i % kFold != 0]
    # Create Subset datasets
    trainDataset = Subset(dataset, trainIndices)
    testDataset = Subset(dataset, testIndices)
    # Create DataLoaders
    trainLoader = DataLoader(
        trainDataset,
        batch_size=batchSize,
        shuffle=True,
        num_workers=numWorkers
    )
    testLoader = DataLoader(
        testDataset,
        batch_size=batchSize,
        shuffle=False,
        num_workers=numWorkers
    )
    return trainLoader, testLoader


if __name__ == "__main__":
    from transformers import Wav2Vec2ForCTC, Wav2Vec2Config, Wav2Vec2Processor
    from fcbh_dataset import *
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2ENGWEB.db"
    audioPath = os.getenv("FCBH_DATASET_FILES") + "/ENGWEB/ENGWEBN2DA"
    model_name = "facebook/mms-1b-all"
    wav2Vec2Processor = Wav2Vec2Processor.from_pretrained(model_name)
    data = FCBHDataset(dbPath, audioPath, wav2Vec2Processor)
    trainDS, testDS = FCBHDataLoader(data, 5, 1, 1)
    print("train", trainDS)
    print("test", testDS)
    data.Close()