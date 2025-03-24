package tests

import (
	"testing"
)

const databaseS3 = `is_new: no
dataset_name: 4a_database_s3_ENGWEB
bible_id: ENGWEB
username: GaryNTest
output:
  sqlite: yes
database:
  aws_s3: s3://dataset-io/GaryNTest/01a_plain_text_ENGWEB/00004/database/01a_plain_text_ENGWEB.db
`

func TestDatabaseS3Direct(t *testing.T) {
	var tests []SqliteTest
	tests = append(tests, SqliteTest{"SELECT count(*) FROM scripts", 7958})
	DirectSqlTest(databaseS3, tests, t)
}
