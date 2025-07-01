from sqlite_utility import *

"""
This file reduces the data being trained to exclude data that likely includes errors.
The fcbh-dataset-io program, also called Artificial Polyglot is being used to proof
the correctness of audio files, but there is a fundamental problem with this that
the training might learn the errors, and therefore not be able to identify them as errors.
To mitigate this problem, this step is using data found during forced alignment to remove
verses that are likely to have errors.
The forced alignment is returning a pair of timestamps and a probability of correctness for each
character.  I summarize these probabilities of correctness to an average for each word,
and each script line.
It is my experience that most characters and words that are missing from the audio or in the
wrong sequence have a probability assigned to them of less that 0.0001.  To make sure
that all errors are removed, I eliminate more lines that just those.

"""

def dataPruner(database):
    query = """CREATE TEMPORARY TABLE pruned_data AS
                SELECT script_id
                FROM scripts WHERE verse_str != '0'
                AND fa_score > 0.5
                AND script_id NOT IN
                (SELECT script_id FROM words WHERE ttype='W' AND fa_score < 0.2)
                """
    database.select(query,())


if __name__ == "__main__":
    dbPath = os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2CUL_MNT.db"
    database = SqliteUtility(dbPath)
    dataPruner(database)
    count = database.selectOne("SELECT count(*) FROM quality_lines",())[0]
    print(count)

# Total records 8122
# Remove verse 0 -> 7862
# Remove verses with word error < 0.00001 -> 7860
# Remove verses with word error < 0.0001 -> 7817
# Remove verses with word error < 0.001 -> 7793
# Remove verses with word error < 0.01 -> 7757
# Remove verses with word error < 0.1 -> 7757
# Remove verses with word error < 0.2 -> 7305
# Remove verses with word error < 0.3 -> 5135
# Remove verses with word error < 0.4 -> 3333
# Remove verses with word error < 0.5 -> 850




