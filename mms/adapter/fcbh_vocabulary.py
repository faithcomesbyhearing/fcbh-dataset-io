import os
import unicodedata
import json
from sqlite_utility import *


def getFCBHVocabulary(databasePath):
    database = SqliteUtility(databasePath)
    chars = set()
    #words = database.select("SELECT word FROM words WHERE ttype='W'", ())
    texts = database.select("SELECT script_text FROM scripts", ())
    for vs in texts:
        line = unicodedata.normalize("NFC", vs[0])
        for ch in line:
            chars.add(ch)
    # Possibly excluding or including hyphens should be a language option.
    #chars.discard('\u002d') # hyphen
    #chars.discard('\u2014') # another hyphen
    #chars.discard('\u2019') # right single quote
    chars = sorted(chars)
    result = {}
    result["<pad>"] = 0
    result["<s>"] = 1
    result["</s>"] = 2
    result["<unk>"] = 3
    result["|"] = 4
    #result[" "] = 5
    index = 5
    for ch in chars:
        result[ch] = index
        index += 1
    filePath = "vocab.json"
    with open(filePath, 'w', encoding='utf-8') as vocabFile:
        json.dump(result, vocabFile)
    return filePath, result

if __name__ == "__main__":
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2ENGWEB.db"
    vocabFile, result = getFCBHVocabulary(dbPath)
    with open(vocabFile, 'r', encoding='utf-8') as file:
        data = json.load(file)
    print(data)
