#!/bin/bash

# Script to generate init.db data from MySQL database
# Usage: ./generate_data_script.sh [FILESET_ID] [BOOK_ID] [OUTPUT_FILE]
# Example: ./generate_data_script.sh ENGNIVN1DA JHN engnivn1da_jhn_data.sql

set -e

# Default values
FILESET_ID=${1:-"ENGNIVN1DA"}
BOOK_ID=${2:-""}
OUTPUT_FILE=${3:-"${FILESET_ID,,}_${BOOK_ID,,}_data.sql"}

echo "Generating data script for fileset: $FILESET_ID"
if [ ! -z "$BOOK_ID" ]; then
    echo "Filtering for book: $BOOK_ID"
fi
echo "Output file: $OUTPUT_FILE"
echo ""

# Check if DBP_MYSQL_DSN is set
if [ -z "$DBP_MYSQL_DSN" ]; then
    echo "Error: DBP_MYSQL_DSN environment variable not set"
    echo "Please run: source ../../setup_env.sh"
    exit 1
fi

# Create the output file
cat > "$OUTPUT_FILE" << EOF
-- ${FILESET_ID} fileset data for init.db
-- Generated from MySQL database on $(date)

-- Fileset data
EOF

echo "Extracting fileset data..."
mysql -u root -e "USE jrs; 
SELECT CONCAT('INSERT INTO bible_filesets (id, hash_id, asset_id, set_type_code, set_size_code, hidden, content_loaded, archived) VALUES (',
    QUOTE(id), ', ',
    QUOTE(hash_id), ', ',
    QUOTE(asset_id), ', ',
    QUOTE(set_type_code), ', ',
    QUOTE(set_size_code), ', ',
    hidden, ', ',
    content_loaded, ', ',
    archived, ');')
FROM bible_filesets 
WHERE id = '$FILESET_ID';" --batch --raw --skip-column-names >> "$OUTPUT_FILE"

echo "" >> "$OUTPUT_FILE"
echo "-- File data (first 6 files for testing)" >> "$OUTPUT_FILE"

echo "Extracting file data..."
if [ ! -z "$BOOK_ID" ]; then
    # Filter by specific book
    mysql -u root -e "USE jrs; 
    SELECT CONCAT('INSERT INTO bible_files (id, hash_id, book_id, chapter_start, file_name, file_size, duration) VALUES (',
        id, ', ',
        QUOTE(hash_id), ', ',
        QUOTE(book_id), ', ',
        chapter_start, ', ',
        QUOTE(file_name), ', ',
        COALESCE(file_size, 'NULL'), ', ',
        COALESCE(duration, 'NULL'), ');')
    FROM bible_files 
    WHERE hash_id = (SELECT hash_id FROM bible_filesets WHERE id = '$FILESET_ID')
    AND book_id = '$BOOK_ID'
    ORDER BY chapter_start 
    LIMIT 6;" --batch --raw --skip-column-names >> "$OUTPUT_FILE"
else
    # Get first 6 files from any book
    mysql -u root -e "USE jrs; 
    SELECT CONCAT('INSERT INTO bible_files (id, hash_id, book_id, chapter_start, file_name, file_size, duration) VALUES (',
        id, ', ',
        QUOTE(hash_id), ', ',
        QUOTE(book_id), ', ',
        chapter_start, ', ',
        QUOTE(file_name), ', ',
        COALESCE(file_size, 'NULL'), ', ',
        COALESCE(duration, 'NULL'), ');')
    FROM bible_files 
    WHERE hash_id = (SELECT hash_id FROM bible_filesets WHERE id = '$FILESET_ID')
    ORDER BY book_id, chapter_start 
    LIMIT 6;" --batch --raw --skip-column-names >> "$OUTPUT_FILE"
fi

echo "" >> "$OUTPUT_FILE"
echo "-- Timestamp data (for first file only)" >> "$OUTPUT_FILE"

echo "Extracting timestamp data..."
if [ ! -z "$BOOK_ID" ]; then
    # Get timestamps for first file of specific book
    mysql -u root -e "USE jrs; 
    SELECT CONCAT('INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (',
        id, ', ',
        bible_file_id, ', ',
        QUOTE(verse_start), ', ',
        COALESCE(QUOTE(verse_end), 'NULL'), ', ',
        timestamp, ', ',
        COALESCE(timestamp_end, 'NULL'), ', ',
        COALESCE(verse_sequence, 'NULL'), ');')
    FROM bible_file_timestamps 
    WHERE bible_file_id = (
        SELECT id FROM bible_files 
        WHERE hash_id = (SELECT hash_id FROM bible_filesets WHERE id = '$FILESET_ID')
        AND book_id = '$BOOK_ID'
        ORDER BY chapter_start 
        LIMIT 1
    )
    ORDER BY verse_sequence;" --batch --raw --skip-column-names >> "$OUTPUT_FILE"
else
    # Get timestamps for first file of any book
    mysql -u root -e "USE jrs; 
    SELECT CONCAT('INSERT INTO bible_file_timestamps (id, bible_file_id, verse_start, verse_end, timestamp, timestamp_end, verse_sequence) VALUES (',
        id, ', ',
        bible_file_id, ', ',
        QUOTE(verse_start), ', ',
        COALESCE(QUOTE(verse_end), 'NULL'), ', ',
        timestamp, ', ',
        COALESCE(timestamp_end, 'NULL'), ', ',
        COALESCE(verse_sequence, 'NULL'), ');')
    FROM bible_file_timestamps 
    WHERE bible_file_id = (
        SELECT id FROM bible_files 
        WHERE hash_id = (SELECT hash_id FROM bible_filesets WHERE id = '$FILESET_ID')
        ORDER BY book_id, chapter_start 
        LIMIT 1
    )
    ORDER BY verse_sequence;" --batch --raw --skip-column-names >> "$OUTPUT_FILE"
fi


echo ""
echo "✅ Data script generated successfully: $OUTPUT_FILE"
echo ""
echo "To use this script:"
echo "1. Review the generated file: cat $OUTPUT_FILE"
echo "2. Update create_init_db.sh to use this file"
echo "3. Run: ./create_init_db.sh"
echo ""
echo "Generated file contains:"
echo "- Fileset data for $FILESET_ID"
if [ ! -z "$BOOK_ID" ]; then
    echo "- First 6 files from book $BOOK_ID"
    echo "- All timestamps for the first file of $BOOK_ID"
else
    echo "- First 6 files from the fileset"
    echo "- All timestamps for the first file"
fi
