#!/bin/bash

# Script to create init.db with simplified schema for HLS stream testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCHEMA_FILE="$SCRIPT_DIR/init_schema.sql"
DB_FILE="$SCRIPT_DIR/init.db"

echo "Creating init.db database..."

# Remove existing database if it exists
if [ -f "$DB_FILE" ]; then
    echo "Removing existing $DB_FILE"
    rm "$DB_FILE"
fi

# Create the database from schema
echo "Creating database from schema..."
sqlite3 "$DB_FILE" < "$SCHEMA_FILE"

# Insert ENGNIVN1DA test data
DATA_FILE="$SCRIPT_DIR/engnivn1da_jhn_data.sql"
if [ -f "$DATA_FILE" ]; then
    echo "Inserting ENGNIVN1DA JHN test data..."
    sqlite3 "$DB_FILE" < "$DATA_FILE"
    echo "Test data inserted successfully!"
else
    echo "Warning: $DATA_FILE not found, creating empty database"
fi

# Verify the database was created and show table info
echo "Database created successfully!"
echo "Tables in $DB_FILE:"
sqlite3 "$DB_FILE" ".tables"

echo "Data verification:"
echo "Filesets:"
sqlite3 "$DB_FILE" "SELECT id, hash_id, set_type_code FROM bible_filesets;"
echo "Files:"
sqlite3 "$DB_FILE" "SELECT id, book_id, chapter_start, file_name FROM bible_files LIMIT 3;"
echo "Timestamps:"
sqlite3 "$DB_FILE" "SELECT id, verse_start, timestamp FROM bible_file_timestamps LIMIT 3;"
echo "Stream bandwidths:"
sqlite3 "$DB_FILE" "SELECT id, file_name, bandwidth FROM bible_file_stream_bandwidths;"
echo "Stream bytes:"
sqlite3 "$DB_FILE" "SELECT id, runtime, bytes, offset FROM bible_file_stream_bytes;"

echo "init.db is ready for HLS stream testing!"
