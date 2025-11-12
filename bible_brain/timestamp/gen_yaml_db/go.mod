module github.com/faithcomesbyhearing/fcbh-dataset-io/bible_brain/timestamp/gen_yaml_db

go 1.23.0

toolchain go1.24.5

require (
	github.com/faithcomesbyhearing/fcbh-dataset-io v0.0.0
	github.com/go-sql-driver/mysql v1.8.1
)

require filippo.io/edwards25519 v1.1.0 // indirect

replace github.com/faithcomesbyhearing/fcbh-dataset-io => ../../..
