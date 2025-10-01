# DBP Test Database Setup

This directory contains scripts and data for creating a local SQLite test database (`init.db`) for testing DBP (Digital Bible Platform) functionality, specifically HLS stream creation and timestamp management.

## Quick Start

To create the test database with John (JHN) data:

```bash
# Create init.db from existing data
./create_init_db.sh

# Verify database contents
sqlite3 init.db "SELECT 'Filesets:' as table_name, COUNT(*) as count FROM bible_filesets UNION ALL SELECT 'Files:', COUNT(*) FROM bible_files UNION ALL SELECT 'Timestamps:', COUNT(*) FROM bible_file_timestamps;"
```

The database will contain:
- 1 fileset (ENGNIVN1DA)
- 6 John files (chapters 1-6) 
- 52 timestamps (complete John chapter 1 timing)
- Empty HLS stream tables (ready for testing)

---

## Detailed Documentation

### Overview

The test database is designed to:
- Test insert timestamp functionality
- Test HLS stream creation and byte-based streaming
- Provide focused test data for specific books (e.g., John chapters)
- Work alongside existing `engnivn1da_timings.db` timing data

### Files

#### Core Scripts
- **`create_init_db.sh`** - Main script to create `init.db` from scratch
- **`generate_data_script.sh`** - Generates SQL data files from MySQL database
- **`init_schema.sql`** - SQLite schema for the test database

#### Data Files
- **`engnivn1da_jhn_data.sql`** - Sample data for ENGNIVN1DA fileset, John chapters 1-6 (committed to git)
- **`init.db`** - SQLite test database (created by scripts, not committed)

### Detailed Usage

### Generate Data for Different Books

```bash
# Generate data for John (JHN)
./generate_data_script.sh ENGNIVN1DA JHN engnivn1da_jhn_data.sql

# Generate data for Matthew (MAT)
./generate_data_script.sh ENGNIVN1DA MAT engnivn1da_mat_data.sql

# Generate data for any book (first 6 files)
./generate_data_script.sh ENGNIVN1DA "" engnivn1da_all_data.sql
```

### Customize Database Creation

1. **Generate new data file**:
   ```bash
   ./generate_data_script.sh [FILESET_ID] [BOOK_ID] [OUTPUT_FILE]
   ```

2. **Update create_init_db.sh** to use your data file:
   ```bash
   # Edit create_init_db.sh and change DATA_FILE variable
   DATA_FILE="$SCRIPT_DIR/your_data_file.sql"
   ```

3. **Recreate database**:
   ```bash
   ./create_init_db.sh
   ```

### Database Schema

The `init.db` database contains these tables:

### `bible_filesets`
- Fileset metadata (id, hash_id, asset_id, set_type_code, etc.)
- Simplified schema without foreign key constraints
- Nullable fields for set_type_code, set_size_code, asset_id

### `bible_files`
- File metadata (id, hash_id, book_id, chapter_start, file_name, etc.)
- Links to filesets via hash_id
- Contains audio file information

### `bible_file_timestamps`
- Verse timing data (verse_start, verse_end, timestamp, etc.)
- Links to files via bible_file_id
- Complete timing data for testing

### `bible_file_stream_bandwidths`
- HLS stream bandwidth information
- Links to files via bible_file_id
- Contains codec and resolution data

### `bible_file_stream_bytes`
- Byte-based streaming data
- Links to stream bandwidths and timestamps
- Contains runtime, bytes, and offset information

### Sample Data

The default `init.db` contains:

- **1 Fileset**: ENGNIVN1DA (audio, NT)
- **6 Files**: John chapters 1-6 (JHN)
- **52 Timestamps**: Complete verse timing for John chapter 1
- **0 Stream Bandwidths**: Empty (ready for HLS stream creation)
- **0 Stream Bytes**: Empty (ready for HLS stream creation)

### Integration with Existing Data

This test database is designed to work alongside your existing `engnivn1da_timings.db` which contains:
- Scripts and idents data
- Additional timing data for other books
- Cross-references for validation

### Troubleshooting

### Environment Issues
```bash
# Check if DBP_MYSQL_DSN is set
echo $DBP_MYSQL_DSN

# If not set, source the environment
source ../../../../setup_env.sh
```

### Database Issues
```bash
# Remove and recreate database
rm init.db
./create_init_db.sh
```

### Data Generation Issues
```bash
# Check MySQL connection
mysql -u root -e "USE jrs; SELECT COUNT(*) FROM bible_filesets;"

# Verify fileset exists
mysql -u root -e "USE jrs; SELECT id, hash_id FROM bible_filesets WHERE id = 'ENGNIVN1DA';"
```

### Development Notes

- The schema is simplified for testing - foreign key constraints are removed
- Data is focused on specific books to align with available timing data
- HLS stream tables are empty by default - ready for testing insert functionality
- The database is designed to be easily recreated and customized

### Next Steps

1. **Test insert functionality** with the populated database
2. **Generate data for other books** as needed
3. **Integrate with existing test suites** for comprehensive testing
4. **Extend schema** if additional testing requirements arise
