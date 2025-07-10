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
        batches = self.database.select('SELECT idx, memory_mb, indexes FROM batches ORDER BY RANDOM()', ())
        for (index, memoryMB, indexes) in batches:
            print("Batch", index, memoryMB, "MB")
            batch = [int(x) for x in indexes.split(',')]
            yield batch

if __name__ == "__main__":
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2ENGWEB.db"
    database = SqliteUtility(dbPath)
    batches = MySampler(database)
    index = 0
    for batch in batches:
        print(index, len(batch), batch)
        index += 1
