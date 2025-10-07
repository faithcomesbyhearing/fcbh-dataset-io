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
    if len(tensor) > originalLen:
        mask[originalLen:] = 0  # Mark padding as invalid
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