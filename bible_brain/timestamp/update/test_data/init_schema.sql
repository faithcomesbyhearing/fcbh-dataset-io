-- Simplified schema for init.db - ready to receive HLS stream inserts
-- Based on dbp_TEST_schema_20250121.sql with modifications

-- Table: bible_filesets
CREATE TABLE bible_filesets (
  id TEXT NOT NULL,
  hash_id TEXT NOT NULL,
  asset_id TEXT NULL,
  set_type_code TEXT NULL,
  set_size_code TEXT NULL,
  hidden INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  content_loaded INTEGER NOT NULL DEFAULT 0,
  archived INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (hash_id),
  UNIQUE (id, asset_id, set_type_code)
);

-- Table: bible_files
CREATE TABLE bible_files (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  hash_id TEXT NOT NULL,
  book_id TEXT NULL,
  chapter_start INTEGER NULL,
  chapter_end INTEGER NULL,
  verse_start TEXT NULL,
  verse_end TEXT NULL,
  file_name TEXT NOT NULL,
  file_size INTEGER NULL,
  duration INTEGER NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  verse_sequence INTEGER NULL,
  UNIQUE (hash_id, book_id, chapter_start, verse_start),
  FOREIGN KEY (hash_id) REFERENCES bible_filesets (hash_id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Table: bible_file_timestamps
CREATE TABLE bible_file_timestamps (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  bible_file_id INTEGER NOT NULL,
  verse_start TEXT NULL,
  verse_end TEXT NULL,
  timestamp REAL NOT NULL,
  timestamp_end REAL NULL,
  verse_sequence INTEGER NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (bible_file_id) REFERENCES bible_files (id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Table: bible_file_stream_bandwidths
CREATE TABLE bible_file_stream_bandwidths (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  bible_file_id INTEGER NOT NULL,
  file_name TEXT NOT NULL,
  bandwidth INTEGER NOT NULL,
  resolution_width INTEGER NULL,
  resolution_height INTEGER NULL,
  codec TEXT NOT NULL DEFAULT '',
  stream INTEGER NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (bible_file_id) REFERENCES bible_files (id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Table: bible_file_stream_bytes
CREATE TABLE bible_file_stream_bytes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  stream_bandwidth_id INTEGER NOT NULL,
  runtime REAL NOT NULL,
  bytes INTEGER NOT NULL,
  offset INTEGER NOT NULL,
  timestamp_id INTEGER NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (stream_bandwidth_id) REFERENCES bible_file_stream_bandwidths (id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (timestamp_id) REFERENCES bible_file_timestamps (id) ON DELETE CASCADE ON UPDATE CASCADE
);
