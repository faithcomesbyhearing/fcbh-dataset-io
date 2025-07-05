# SqliteUtility.py
#
# This program provides convenience methods wrapping the Sqlite3 client.
# This class supports the same interface as SQLUtility whenever possible
#

import os
import sys
import sqlite3

class SqliteUtility:

    def __init__(self, databasePath):
        self.conn = sqlite3.connect(databasePath)

    def close(self):
        if self.conn != None:
            self.conn.close()
            self.conn = None

    def selectOne(self, statement, values):
        cursor = self.conn.cursor()
        try:
            cursor.execute(statement, values)
            result = cursor.fetchone()
            cursor.close()
            return result
        except Exception as err:
            self.error(cursor, statement, err)

    def select(self, statement, values):
        cursor = self.conn.cursor()
        try:
            cursor.execute(statement, values)
            resultSet = cursor.fetchall()
            cursor.close()
            return resultSet
        except Exception as err:
            self.error(cursor, statement, err)

    def execute(self, statement, values):
        cursor = conn.cursor()
        try:
            cursor.execute(statement, values)
            cursor.close()
            conn.commit()
         except Exception as err:
             self.error(cursor, statement, err)

    def error(self, cursor, stmt, error):
        cursor.close()
        print("ERROR executing SQL %s on '%s'" % (error, stmt))
        self.conn.rollback()
        sys.exit(1)


if __name__ == "__main__":
    sql = SqliteUtility(os.getenv("FCBH_DATASET_DB") + "/GaryNTest/N2ENGWEB.db")
    count = sql.selectOne("select count(*) from scripts",())
    print(count)
    (audioFile, text) = sql.selectOne('select audio_file, script_text from scripts where script_id = ?', (10,))
    print(audioFile, text)
    sql.close()
