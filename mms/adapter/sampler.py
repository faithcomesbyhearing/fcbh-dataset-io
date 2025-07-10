from torch.utils.data import Sampler
from sqlite_utility import *

class MySampler(Sampler):
    def __init__(self, database):
        super().__init__()
        self.database = database

    def __len__(self):
      count = self.database.selectOne('SELECT count(*) FROM batches', ())
      return count[0]

    def __iter__(self):
        batches = self.database.select('SELECT idx, num_samples, memory_mb, padded_mb, indexes FROM batches', ())
        for (index, numSamples, memoryMB, paddedMB, indexes) in batches:
            print("Batch", index, numSamples, memoryMB, "MB", paddedMB, "MB")
            batch = [int(x) for x in indexes.split(',')]
            yield batch

if __name__ == "__main__":
    #dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2ENGWEB.db"
    dbPath = os.getenv("FCBH_DATASET_TMP") + "/N2KEUWB4.db"
    database = SqliteUtility(dbPath)
    batches = MySampler(database)
    index = 0
    for batch in batches:
        print(index, len(batch), batch)
        index += 1
