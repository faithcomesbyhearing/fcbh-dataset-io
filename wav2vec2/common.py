import sys
import numpy as np
import torch
import blosc


def ensureMinimumTensorSize(tensor, minTensorLength, padValue):
    originalLen = tensor.shape[-1]
    if originalLen < minTensorLength:
        paddingNeeded = minTensorLength - originalLen
        tensor = torch.nn.functional.pad(tensor, (0, paddingNeeded), mode='constant', value=padValue)
    mask = torch.ones(len(tensor), dtype=torch.int)
    #if len(tensor) > originalLen:
    #    mask[originalLen:] = 0  # Mark padding as invalid
    return tensor, mask


def compress(tensor):
    numpy_array = tensor.cpu().numpy()
    compressed = blosc.compress(
        numpy_array.tobytes(),
        typesize = numpy_array.itemsize,
        cname = 'lz4',
        clevel = 5,
        shuffle = blosc.SHUFFLE
    )
    return compressed


def decompress(blob, dtype, numSamples=1):
    decompressed = blosc.decompress(blob)
    numpy_array = np.frombuffer(decompressed, dtype=dtype).copy()
    if numSamples != 1:
        if len(numpy_array) % numSamples != 0:
            print("Cannot reshape", len(numpy_array), "element into", numSamples, file=sys.stderr)
        numpy_array = numpy_array.reshape(numSamples, -1)
    tensor = torch.from_numpy(numpy_array)
    return tensor


if __name__ == "__main__":
    import os
    sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'train'))
    from sqlite_utility import *
    from dataset import *
    dbPath = os.getenv("FCBH_DATASET_TMP") + "/N2KEUWB4_CTC.db"
    database = SqliteUtility(dbPath)
    query = 'SELECT num_samples, input_values, attention_mask, labels, memory_mb FROM tensors'
    batches = database.select(query, ())
    for (numSamples, inputValuesBlob, attentionMaskBlob, labelsBlob, memoryMB) in batches:
        print("numSamples", numSamples)
        inputValues = decompress(inputValuesBlob, np.float32, numSamples)
        inputValuesBlob2 = compress(inputValues)
        inputValues2 = decompress(inputValuesBlob2, np.float32, numSamples)
        if not torch.equal(inputValues, inputValues2):
            print("MISMATCH")
            sys.exit(1)
        else:
            print("INPUT EQUAL")
        mask = decompress(attentionMaskBlob, np.int32, numSamples)
        maskBlob2 = compress(mask)
        mask2 = decompress(maskBlob2, np.int32, numSamples)
        if not torch.equal(mask, mask2):
            print("MISMATCH MASK")
            sys.exit(1)
        labels = decompress(labelsBlob, np.int64, numSamples)
        labelsBlob2 = compress(labels)
        labels2 = decompress(labelsBlob2, np.int64, numSamples)
        if not torch.equal(labels, labels2):
            print("MISMATCH LABELS")
            sys.exit(1)
