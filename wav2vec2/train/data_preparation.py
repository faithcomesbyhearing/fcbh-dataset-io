import os
import sys
import torch
import torchaudio
import numpy as np
from transformers import Wav2Vec2Processor
from sqlite_utility import *
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
from common import ensureMinimumTensorSize, compress, decompress
import time
import blosc

"""
NOTE: def dataPruner has a bug.  In a recent run it pruned out ouly 214 verses when it was
given an input of 0.1, which should have eliminated 10%.
As of 10/28/75, this model is not being used.  So, I have not bothered to fix it.
"""

# This file does all of the data preparation prior to training, and its content is also used for ASR
# Each step performs a process and stores the result in a sqlite database in the TMP directory.
# def prepareDataset() creates a table called samples that contains each word sample in the audio
# the data is a tensor that has been padded to a minimum size of too small, and compressed.
# def identifyBatches identifies the samples that go together in a batch and creates a record
# in sqlite of the samples that go in each batch.
# def prepareBatches() creates the actual tensors, and stores them compressed in a sqlite table.


def dataPreparation(scriptsDB, scriptsDBPath, audioDir, processor, maxBatchSize, targetMemoryMB, minAudioSec,
    pctExclude):
    (minTensorLength, minTensorMB) = calcMinTensor(minAudioSec, torch.float32)
    print("minAudioSec:", minAudioSec, "minTensorLength:", minTensorLength, "minTensorMB:", minTensorMB)
    scriptsDBName = os.path.basename(scriptsDBPath)
    samplesDBPath = os.path.join(os.getenv('FCBH_DATASET_TMP'), scriptsDBName)
    if os.path.exists(samplesDBPath):
        os.remove(samplesDBPath)
    samplesDB = SqliteUtility(samplesDBPath)
    numSamples = prepareDataset(scriptsDB, samplesDB, audioDir, processor) # 22 sec
    print("numSamples", numSamples)
    faErrorCutoff = dataPruner(scriptsDB, pctExclude)
    print("faErrorCutoff", faErrorCutoff)
    numBatches = identifyBatches(samplesDB, maxBatchSize, targetMemoryMB, minTensorMB, faErrorCutoff)
    print("numBatches", numBatches)
    numTensors = prepareBatches(samplesDB, processor, minTensorLength)
    print("numTensors", numTensors)
    return samplesDB


def calcMinTensor(minAudioSec, dtype):
    minTensorLength = minAudioSec * 16000
    bytesPerElement = torch.finfo(dtype).bits // 8
    minTensorMB = (minTensorLength * bytesPerElement) / (1024 * 1024)
    return (int(minTensorLength), minTensorMB)


"""
The fcbh-dataset-io program, also called Artificial Polyglot is being used to proof
the correctness of audio files, but there is a fundamental problem with this that
the training might learn the errors, and therefore not be able to identify them as errors.
To mitigate this problem, this step is using data found during forced alignment to remove
words that are likely to have errors.
The forced alignment is returning a pair of timestamps and a probability of correctness for each
character.  I summarize these probabilities of correctness to an average for each word,
and each script line.
"""

def dataPruner(database, pctExclude):
    list = database.select("SELECT fa_score FROM words WHERE ttype='W' ORDER by fa_score",())
    cutoffPos = int(len(list) * pctExclude)
    faErrorCutoff = list[cutoffPos][0]
    return faErrorCutoff


def prepareDataset(scriptsDB, samplesDB, audioDir, processor):
    samplesDB.execute('DROP TABLE IF EXISTS samples', ())
    samples = """CREATE TABLE samples (idx INTEGER PRIMARY KEY, script_id INTEGER, word_id INTEGER,
            input_values BLOB, labels BLOB, text TEXT, reference TEXT, fa_score FLOAT, memory_mb FLOAT)"""
    samplesDB.execute(samples,())
    query = """
        SELECT s.script_id, w.word_id, s.book_id || ' ' || s.chapter_num || ':' || s.verse_str || '.' || w.word_seq as ref,
                s.audio_file, w.word_begin_ts, w.word_end_ts, w.fa_score, w.word AS text
        FROM scripts s
        JOIN words w ON w.script_id = s.script_id
        WHERE w.ttype = 'W'
        ORDER BY w.word_end_ts - w.word_begin_ts -- sort by size
        """
    index = -1
    data = scriptsDB.select(query,())
    for (scriptId, wordId, reference, audioFile, beginTS, endTS, faScore, text) in data:
        index += 1
        audioFile = audioFile.replace(".mp3", ".wav")
        audioPath = os.path.join(audioDir, audioFile)

        info = torchaudio.info(audioPath, format="wav")
        if info.sample_rate != 16000:
            print("Audio sample rate must be 16000", file=sys.stderr, flush=True)
            sys.exit(1)

        speech, sample_rate = torchaudio.load(
            audioPath,
            frame_offset = int(beginTS * 16000),
            num_frames = int((endTS - beginTS) * 16000)
        )
        speech = speech.squeeze().numpy()

        inputValues = processor(
                speech,
                sampling_rate=16000,
                return_tensors=None,
                padding=False
            ).input_values
        inputValues = np.array(inputValues)
        inputValuesTensor = torch.tensor(inputValues, dtype=torch.float).squeeze(0)
        inputValuesBlob = compress(inputValuesTensor)
        memoryMB = inputValuesTensor.element_size() * inputValuesTensor.numel() / (1024 * 1024)
        if memoryMB == 0:
            print(reference, audioFile, "Has Zero Length input_values", file=sys.stderr, flush=True)
            sys.exit(1)

        labels = processor(text=text.lower()).input_ids
        labelsTensor = torch.tensor(labels, dtype=torch.long)
        labelsBlob = compress(labelsTensor)

        insert = """INSERT INTO samples (idx, script_id, word_id, input_values, labels, text, reference,
            fa_score, memory_mb) VALUES (?,?,?,?,?,?,?,?,?)"""
        samplesDB.execute(insert, (index, scriptId, wordId, inputValuesBlob, labelsBlob, text, reference, faScore, memoryMB))
    return index + 1


def identifyBatches(database, maxBatchSize, targetMemoryMB, minTensorMB, faErrorCutoff):
    database.execute('DROP TABLE IF EXISTS batches', ())
    batches = """CREATE TABLE batches (idx INTEGER PRIMARY KEY, num_samples INTEGER, memory_mb FLOAT,
                padded_mb FLOAT, indexes BLOB)"""
    database.execute(batches,())
    query = 'SELECT idx, text, reference, memory_mb FROM samples WHERE fa_score > ?'
    samples = database.select(query, (faErrorCutoff,))
    insert = 'INSERT INTO batches (idx, num_samples, memory_mb, padded_mb, indexes) VALUES (?,?,?,?,?)'
    batchNum = 0
    batch = []
    unpaddedSize = 0.0
    paddedSize = 0.0
    for (index, text, reference, memoryMB) in samples:
        memoryMB = max(memoryMB, minTensorMB)
        if len(batch) >= maxBatchSize or (memoryMB * (len(batch) + 1)) >= targetMemoryMB:
            database.execute(insert, (batchNum, len(batch), unpaddedSize, paddedSize,
                ','.join(map(str, batch))))
            batchNum += 1
            batch = []
            unpaddedSize = 0.0
            paddedSize = 0.0
        batch.append(index)
        unpaddedSize += memoryMB
        paddedSize = len(batch) * memoryMB
    if len(batch) > 0:
        database.execute(insert, (batchNum, len(batch), unpaddedSize, paddedSize,
            ','.join(map(str, batch))))
    return batchNum + 1


def prepareBatches(database, processor, minTensorLength):
    database.execute('DROP TABLE IF EXISTS tensors', ())
    padding: Union[bool, str] = True
    table = """CREATE TABLE tensors (idx INTEGER PRIMARY KEY, num_samples INTEGER, memory_mb FLOAT,
                input_values BLOB, attention_mask BLOB, labels BLOB)"""
    database.execute(table, ())
    batches = database.select('SELECT idx, num_samples, memory_mb, padded_mb, indexes FROM batches', ())
    for (index, numSamples, memoryMB, paddedMB, indexes) in batches:
        #print("\nIndex", index, numSamples, memoryMB, paddedMB, indexes)
        samples = indexes.split(',')
        inputFeatures = []
        labelFeatures = []
        for s in samples:
            sampleQuery = """SELECT idx, input_values, labels, text, reference, memory_mb
                            FROM samples WHERE idx = ?"""
            (sampleIdx, inputValues, labels, text, reference, memoryMB) = database.selectOne(sampleQuery, (int(s),))
            inputTensor = decompress(inputValues, np.float32)
            originalLen = len(inputTensor)
            inputTensor, attentionMask = ensureMinimumTensorSize(inputTensor, minTensorLength, 0)
            inputFeatures.append({
                "input_values": inputTensor,
                "attention_mask": attentionMask
            })
            labelsTensor = decompress(labels, np.int64)
            labelFeatures.append({"input_ids": labelsTensor})
        batch = processor.pad(
            inputFeatures,
            padding = padding,
            return_tensors = "pt",
        )
        labelsBatch = processor.pad(
            labels = labelFeatures,
            padding = padding,
            return_tensors = "pt",
        )
        # replace padding with -100 to ignore loss correctly
        labels = labelsBatch["input_ids"].masked_fill(labelsBatch.attention_mask.ne(1), -100)
        inputValuesBlob = compress(batch['input_values'])
        attentionMaskBlob = compress(batch['attention_mask'])
        labelsBlob = compress(labels)
        insert = """INSERT INTO tensors (idx, num_samples, memory_mb, input_values,
                            attention_mask, labels) VALUES (?,?,?,?,?,?)"""
        database.execute(insert, (index, numSamples, paddedMB, inputValuesBlob,
                attentionMaskBlob, labelsBlob))
    return len(batches)


def displaySamples(database):
    query = 'SELECT idx, input_values, labels, text, reference, memory_mb FROM samples'
    samples = database.select(query, ())
    print("length", len(samples))
    for (index, inputValues, labels, text, reference, memoryMB) in samples:
        audioTensor = decompress(inputValues, np.float32)
        labelsTensor = decompress(labels, np.int64)
        print("\nIndex", index)
        print("audio", audioTensor.shape, type(audioTensor))
        print("labels", labelsTensor.shape, type(labelsTensor))
        print("text", text)
        print("reference", reference)
        print("memory_mb", memoryMB)
    print("numSamples", len(samples))


def displayBatches(database):
    batches = database.select('SELECT idx, num_samples, memory_mb, padded_mb, indexes FROM batches', ())
    for (index, numSamples, memoryMB, paddedMB, indexes) in batches:
        print("\nIndex", index, numSamples, memoryMB, paddedMB, indexes)
    print("numBatches", len(batches))


def displayTensors(database):
    batches = database.select('SELECT idx, num_samples, memory_mb, input_values, attention_mask, labels FROM tensors', ())
    for (index, numSamples, memoryMB, inputValues, attentionMask, labels) in batches:
        audioTensor = decompress(inputValues, np.float32)
        maskTensor = decompress(attentionMask, np.float32)
        labelsTensor = decompress(labels, np.int64)
        print("\nIndex", index)
        print("audio", audioTensor.shape, type(audioTensor), audioTensor.nbytes / (1024**2))
        print("mask", maskTensor.shape, type(maskTensor), maskTensor.nbytes / (1024**2))
        print("labels", labelsTensor.shape, type(labelsTensor), labelsTensor.nbytes / (1024**2))
        print("memory_mb", memoryMB)


def checkBlob(idx, name, blob, dtype):
    tensor = decompress(blob, dtype)
    if torch.isnan(tensor).any():
        print(f"NaN found in {name} {idx}")
    if torch.isinf(tensor).any():
        print(f"Inf found in {name} {idx}")
    if tensor.numel() == 0:
        print(f"Empty tensor {nme} in sample {idx}")


def checkSamples(samplesDB):
    samples = samplesDB.select('SELECT idx, input_values, labels, text TEXT, reference, memory_mb FROM samples',())
    for (idx, inputValues, labels, text, reference, memoryMB) in samples:
        checkBlob(idx, 'input_values', inputValues, np.float32)
        checkBlob(idx, 'labels', labels, np.int64)


if __name__ == "__main__":
    from tokenizer import createTokenizer
    from transformers import Wav2Vec2Processor, Wav2Vec2FeatureExtractor, Wav2Vec2CTCTokenizer
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2KEUWB4.db"
    database = SqliteUtility(dbPath)
    audioPath = os.getenv("FCBH_DATASET_FILES") + "/N2KEUWB4/N2KEUWBT"
    tokenizer = createTokenizer(database, "keu")
    featureExtractor = Wav2Vec2FeatureExtractor(
        feature_size=1, sampling_rate=16000, padding_value=0.0,
        do_normalize=True, return_attention_mask=True
    )
    processor = Wav2Vec2Processor(feature_extractor=featureExtractor, tokenizer=tokenizer)
    samplesDB = dataPreparation(database, dbPath, audioPath, processor, 1024, 16, 1.0, 0.1)
    database.close()
    #sys.exit()
    displaySamples(samplesDB)
    displayBatches(samplesDB)
    displayTensors(samplesDB)
    checkSamples(samplesDB),
    samplesDB.close()