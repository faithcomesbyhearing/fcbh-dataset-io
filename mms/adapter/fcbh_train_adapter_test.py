

conda activate mms_adapter

#fcbh_train_adapter.py {iso639-3} {vocabSize} {databasePath} {audioDirectory} {batchSize} {numWorkers}

python fcbh_train_adapter.py {iso639-3} {vocabSize} {databasePath} {audioDirectory} {batchSize} {numWorkers}



vocabSize

chars = set()
chars.add(' ')
words = self.database.select("SELECT word FROM words WHERE ttype='W'", ())
for wd in words:
    for ch in wd:
        chars.add(ch)
vocabSize = len(chars)
print("chars", chars)
