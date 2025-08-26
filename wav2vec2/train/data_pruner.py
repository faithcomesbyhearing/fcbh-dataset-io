from sqlite_utility import *

"""
This file reduces the data being trained to exclude data that likely includes errors.
The fcbh-dataset-io program, also called Artificial Polyglot is being used to proof
the correctness of audio files, but there is a fundamental problem with this that
the training might learn the errors, and therefore not be able to identify them as errors.
To mitigate this problem, this step is using data found during forced alignment to remove
words that are likely to have errors.
The forced alignment is returning a pair of timestamps and a probability of correctness for each
character.  I summarize these probabilities of correctness to an average for each word,
and each script line.

The error value, called fa_score is analyzed statistically to find the lowest 10% of the
values, the word_id of the 90% with higher fa_score values are stored in the pruned data table.
"""

def dataPruner(database):
    # Version 1
    database.execute('DROP TABLE IF EXISTS pruned_data',())
    list = database.select("SELECT fa_score FROM words WHERE ttype='W' ORDER by fa_score",())
    tenPctPos = int(len(list)/10)
    tenPct = list[tenPctPos]
    query = """CREATE TEMPORARY TABLE pruned_data AS
            SELECT word_id
            FROM words WHERE ttype = 'W'
            AND fa_score > ?"""
    database.select(query,(tenPct))


if __name__ == "__main__":
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2CUL_MNT.db"
    database = SqliteUtility(dbPath)
    dataPruner(database)
    count = database.selectOne("SELECT count(*) FROM pruned_data",())[0]
    print(count)






