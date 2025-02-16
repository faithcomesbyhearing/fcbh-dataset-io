package update

import "database/sql"

type Timestamp struct {
	TimestampId int64
	VerseStr    string
	VerseEnd    sql.NullString // On update set to null
	VerseSeq    int
	BeginTS     float64
	EndTS       float64
	Duration    float64
	Position    int64
	NumBytes    int64
}

//  `id` int unsigned NOT NULL AUTO_INCREMENT,
//  `bible_file_id` int unsigned NOT NULL,
//  `verse_start` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
//  `verse_end` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
//  `timestamp` float NOT NULL,
//  `timestamp_end` float DEFAULT NULL,
//  `verse_sequence` tinyint unsigned NOT NULL,

// `id` int unsigned NOT NULL AUTO_INCREMENT,
// `stream_bandwidth_id` int unsigned NOT NULL,
// `runtime` double(8,2) NOT NULL,
// `bytes` int NOT NULL,
// `offset` int NOT NULL,
// `timestamp_id` int unsigned NOT NULL,

//  bible_filesets
//  PRIMARY KEY (`hash_id`),
//  UNIQUE KEY `unique_prefix_for_s3` (`id`,`asset_id`,`set_type_code`),
