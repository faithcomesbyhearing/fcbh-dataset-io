import os
import sys
import torch
import torchaudio
import numpy as np
from sqlite_utility import *
from data_pruner import dataPruner


def dataPreparation(database, audioDir, processor):
    exists = database.selectOne("SELECT name FROM sqlite_master WHERE type='table' AND name='samples'", ())
    if exists:
        numScripts = database.selectOne("SELECT count(*) FROM scripts", ())
        numSamples = database.select("SELECT count(*) FROM samples", ())
        if numSamples > (numScripts / 2):
            return numSamples
    samples = 'CREATE TABLE samples (index INTEGER, input_values BLOB, labels BLOB, text TEXT, reference TEXT, memory_mb FLOAT)'
    ### make index a primary key
    database.execute(samples,())
    dataPruner(database)
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
    data = database.select(query,())
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

        labels = processor(text=text).input_ids
        labelsTensor = torch.tensor(labels, dtype=torch.long)

        buffer = BytesIO()
        torch.save(inputValuesTensor, buffer)
        inputValuesBlob = buffer.getvalue()

        buffer = BytesIO()
        torch.save(labelsTensor, buffer)
        labelsBlob = buffer.getvalue()

        update = 'UPDATE samples SET input_values = ?, labels = ?, text = ?, reference = ?, memory_mb = ? WHERE index = ?'
        database.execute(update, (inputValuesBlob, labelsBlob, text, reference, memoryMB, index))


if __name__ == "__main__":
    from tokenizer import createTokenizer
    from transformers import Wav2Vec2Processor, Wav2Vec2FeatureExtractor, Wav2Vec2CTCTokenizer
    from data_pruner import *
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2ENGWEB.db"
    database = SqliteUtility(dbPath)
    audioPath = os.getenv("FCBH_DATASET_FILES") + "/ENGWEB/ENGWEBN2DA"
    tokenizer = createTokenizer(database, "eng")
    featureExtractor = Wav2Vec2FeatureExtractor(
        feature_size=1, sampling_rate=16000, padding_value=0.0,
        do_normalize=True, return_attention_mask=True
    )
    processor = Wav2Vec2Processor(feature_extractor=featureExtractor, tokenizer=tokenizer)
    numSamples = dataPreparation(database, audioPath, processor)
    query = 'SELECT index, input_values, labels, text, reference, memory_mb)'
    dataset = database.select(query, ())
    print("length", len(dataset))
    for (index, inputValues, labels, text, reference, memoryMB) in data:
        buffer = BytesIO(inputValues)
        audioTensor = torch.load(buffer)
        buffer = BytesIO(labels)
        labelsTensor = torch.load(buffer)
        print("\nIndex", index)
        print("audio", audioTensor.shape, type(audioTensor))
        print("labels", labelsTensor.shape, type(labelsTensor), labelsTensor)
        print("text", text)
        print("reference", reference)
        print("memory_mb", memoryMB)
    database.close()