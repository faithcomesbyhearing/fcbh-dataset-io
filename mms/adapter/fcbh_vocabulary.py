import os
import unicodedata
import json
from sqlite_utility import *


def getFCBHVocabulary(databasePath):
    database = SqliteUtility(databasePath)
    chars = set()
    words = database.select("SELECT word FROM words WHERE ttype='W'", ())
    for wd in words:
        word = unicodedata.normalize("NFC", wd[0].lower())
        for ch in word:
            chars.add(ch)
    # Possibly excluding or including hyphens should be a language option.
    chars.discard('\u2014') # another hyphen
    chars.discard('\u002d') # hyphen
    result = {}
    result["<pad>"] = 0
    result["<s>"] = 1
    result["</s>"] = 2
    result["<unk>"] = 3
    result["|"] = 4
    index = 5
    for ch in chars:
        result[ch] = index
        index += 1
    result[" "] = index
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
