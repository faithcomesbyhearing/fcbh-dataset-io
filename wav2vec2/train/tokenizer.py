import os
import unicodedata
import json
from transformers import Wav2Vec2CTCTokenizer
from sqlite_utility import *

#
# https://huggingface.co/blog/mms_adapters
#

def createTokenizer(database, targetLang):
    chars = set()
    texts = database.select("SELECT word FROM words WHERE ttype='W'", ())
    for vs in texts:
        line = unicodedata.normalize("NFC", vs[0].lower())
        for ch in line:
            chars.add(ch)
    //# Possibly excluding or including hyphens should be a language option.
    //chars.discard('\u002d') # hyphen
    //chars.discard('\u2014') # another hyphen
    //chars.discard('\u2019') # right single quote
    chars = sorted(chars)
    vocabDict = {}
    vocabDict["|"] = 0
    index = 1
    for ch in chars:
        vocabDict[ch] = index
        index += 1
    vocabDict["[UNK]"] = len(vocabDict)
    vocabDict["[PAD]"] = len(vocabDict)
    newVocabDict = {targetLang: vocabDict}
    filePath = "vocab.json"
    with open(filePath, 'w', encoding='utf-8') as vocabFile:
        json.dump(newVocabDict, vocabFile)
    tokenizer = Wav2Vec2CTCTokenizer.from_pretrained(
        "./",
        unk_token = "[UNK]",
        pad_token = "[PAD]",
        word_delimiter_token = "|",
        target_lang = targetLang
    )
    return tokenizer


if __name__ == "__main__":
    targetLang = "cul"
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2CUL_MNT.db"
    database = SqliteUtility(dbPath)
    tokenizer = createTokenizer(database, targetLang)
    database.close()
    vocabFile = "vocab.json"
    with open(vocabFile, 'r', encoding='utf-8') as file:
        data = json.load(file)
    print(data)
    text = "Hello, how are you doing today?"
    encoded = tokenizer.encode(text)
    print(f"Encoded token IDs: {encoded}")
    decoded = tokenizer.decode(encoded)
    print(f"Decoded text: {decoded}")
    tokens = tokenizer.tokenize(text)
    print(f"Tokens: {tokens}")