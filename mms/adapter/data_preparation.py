import os
import sys
import torch
import torchaudio
import numpy as np
from transformers import Wav2Vec2Processor
from io import BytesIO
from sqlite_utility import *
from data_pruner import dataPruner
import time


def dataPreparation(scriptsDB, scriptsDBPath, audioDir, processor, maxBatchSize, targetMemoryMB):
    dataPruner(scriptsDB) # 0.02
    scriptsDBName = os.path.basename(scriptsDBPath)
    samplesDBPath = os.path.join(os.getenv('FCBH_DATASET_TMP'), scriptsDBName)
    if os.path.exists(samplesDBPath):
        os.remove(samplesDBPath)
    samplesDB = SqliteUtility(samplesDBPath)
    numSamples = prepareDataset(scriptsDB, samplesDB, audioDir, processor) # 22 sec
    print("numSamples", numSamples)
    numBatches = identifyBatches(samplesDB, maxBatchSize, targetMemoryMB) # 3.8 sec
    print("numBatches", numBatches)
    numTensors = prepareBatches(samplesDB, processor)
    print("numTensors", numTensors)
    return samplesDB


def prepareDataset(scriptsDB, samplesDB, audioDir, processor):
    samplesDB.execute('DROP TABLE IF EXISTS samples', ())
    samples = 'CREATE TABLE samples (idx INTEGER PRIMARY KEY, input_values BLOB, labels BLOB, text TEXT, reference TEXT, memory_mb FLOAT)'
    samplesDB.execute(samples,())
    query = """
        SELECT s.book_id || ' ' || s.chapter_num || ':' || s.verse_str as ref,
            s.audio_file, s.script_begin_ts, s.script_end_ts, GROUP_CONCAT(w.word, ' ') AS text
        FROM scripts s
        JOIN words w ON w.script_id = s.script_id
        WHERE w.ttype = 'W' AND s.script_id IN (SELECT script_id FROM pruned_data)
        GROUP BY s.script_id, s.book_id, s.chapter_num, s.verse_str, s.audio_file, s.script_begin_ts, s.script_end_ts
        ORDER BY s.script_end_ts - s.script_begin_ts
        """
    index = -1
    data = scriptsDB.select(query,())
    for (reference, audioFile, beginTS, endTS, text) in data:
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
        memoryMB = inputValuesTensor.element_size() * inputValuesTensor.numel() / (1024 * 1024)
        if memoryMB == 0:
            print(reference, audioFile, "Has Zero Length input_values", file=sys.stderr, flush=True)
            sys.exit(1)

        labels = processor(text=text.lower()).input_ids
        labelsTensor = torch.tensor(labels, dtype=torch.long)

        buffer = BytesIO()
        torch.save(inputValuesTensor, buffer)
        inputValuesBlob = buffer.getvalue()

        buffer = BytesIO()
        torch.save(labelsTensor, buffer)
        labelsBlob = buffer.getvalue()
        insert = 'INSERT INTO samples (idx, input_values, labels, text, reference, memory_mb) VALUES (?,?,?,?,?,?)'
        samplesDB.execute(insert, (index, inputValuesBlob, labelsBlob, text, reference, memoryMB))
    return index + 1


def identifyBatches(database, maxBatchSize, targetMemoryMB):
    count = database.selectOne('SELECT count(*) FROM samples WHERE memory_mb > ?', (targetMemoryMB,))
    if len(count) > 0 and count[0] > 0:
        print(count[0], "samples were dropped, because they were too large.", file=sys.stderr)
    database.execute('DROP TABLE IF EXISTS batches', ())
    batches = """CREATE TABLE batches (idx INTEGER PRIMARY KEY, num_samples INTEGER, memory_mb FLOAT,
                padded_mb FLOAT, indexes BLOB)"""
    database.execute(batches,())
    query = 'SELECT idx, input_values, labels, text, reference, memory_mb FROM samples WHERE memory_mb <= ?'
    samples = database.select(query, (targetMemoryMB,))
    insert = 'INSERT INTO batches (idx, num_samples, memory_mb, padded_mb, indexes) VALUES (?,?,?,?,?)'
    batchNum = 0
    batch = []
    unpaddedSize = 0.0
    paddedSize = 0.0
    for (index, inputValues, labels, text, reference, memoryMB) in samples:
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


def prepareBatches(database, processor):
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
            buffer = BytesIO(inputValues)
            inputTensor = torch.load(buffer)
            inputFeatures.append({"input_values": inputTensor})
            buffer = BytesIO(labels)
            labelsTensor = torch.load(buffer)
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
        insert = """INSERT INTO tensors (idx, num_samples, memory_mb, input_values,
                    attention_mask, labels) VALUES (?,?,?,?,?,?)"""
        inputValuesBuffer = BytesIO()
        torch.save(batch['input_values'], inputValuesBuffer)
        attentionMaskBuffer = BytesIO()
        torch.save(batch['attention_mask'], attentionMaskBuffer)
        labelsBuffer = BytesIO()
        torch.save(labels, labelsBuffer)
        database.execute(insert, (index, numSamples, paddedMB, inputValuesBuffer.getvalue(),
                attentionMaskBuffer.getvalue(), labelsBuffer.getvalue()))
    return len(batches)


def displaySamples(database):
    query = 'SELECT idx, input_values, labels, text, reference, memory_mb FROM samples'
    samples = database.select(query, ())
    print("length", len(samples))
    for (index, inputValues, labels, text, reference, memoryMB) in samples:
        buffer = BytesIO(inputValues)
        audioTensor = torch.load(buffer)
        buffer = BytesIO(labels)
        labelsTensor = torch.load(buffer)
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
        inputValuesBuffer = BytesIO(inputValues)
        audioTensor = torch.load(inputValuesBuffer)
        maskBuffer = BytesIO(attentionMask)
        maskTensor = torch.load(maskBuffer)
        labelsBuffer = BytesIO(labels)
        labelsTensor = torch.load(labelsBuffer)
        print("\nIndex", index)
        print("audio", audioTensor.shape, type(audioTensor), audioTensor.nbytes / (1024**2))
        print("mask", maskTensor.shape, type(maskTensor), maskTensor.nbytes / (1024**2))
        print("labels", labelsTensor.shape, type(labelsTensor), labelsTensor.nbytes / (1024**2))
        print("memory_mb", memoryMB)


def checkBlob(idx, name, blob):
    buffer = BytesIO(blob)
    tensor = torch.load(buffer)
    if torch.isnan(tensor).any():
        print(f"NaN found in {name} {idx}")
    if torch.isinf(tensor).any():
        print(f"Inf found in {name} {idx}")
    if tensor.numel() == 0:
        print(f"Empty tensor {nme} in sample {idx}")


def checkSamples(samplesDB):
    samples = samplesDB.select('SELECT idx, input_values, labels, text TEXT, reference, memory_mb FROM samples',())
    for (idx, inputValues, labels, text, reference, memoryMB) in samples:
        checkBlob(idx, 'input_values', inputValues)
        checkBlob(idx, 'labels', labels)


if __name__ == "__main__":
    from tokenizer import createTokenizer
    from transformers import Wav2Vec2Processor, Wav2Vec2FeatureExtractor, Wav2Vec2CTCTokenizer
    from data_pruner import *
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2KEUWB4.db"
    database = SqliteUtility(dbPath)
    audioPath = os.getenv("FCBH_DATASET_FILES") + "/N2KEUWB4/N2KEUWBT"
    tokenizer = createTokenizer(database, "keu")
    featureExtractor = Wav2Vec2FeatureExtractor(
        feature_size=1, sampling_rate=16000, padding_value=0.0,
        do_normalize=True, return_attention_mask=True
    )
    processor = Wav2Vec2Processor(feature_extractor=featureExtractor, tokenizer=tokenizer)
    samplesDB = dataPreparation(database, dbPath, audioPath, processor, 128, 32)
    database.close()
    displaySamples(samplesDB)
    displayBatches(samplesDB)
    displayTensors(samplesDB)
    checkSamples(samplesDB)
    samplesDB.close()