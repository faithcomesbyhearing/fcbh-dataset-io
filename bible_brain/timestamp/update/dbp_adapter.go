package update

import (
	"context"
	"database/sql"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

type DBPAdapter struct {
	ctx  context.Context
	conn *sql.DB
}

func NewDBPAdapter(ctx context.Context) (DBPAdapter, *log.Status) {
	var dbp DBPAdapter
	dbp.ctx = ctx
	var err error
	dbp.conn, err = sql.Open("mysql", GetDBPMySqlDSN())
	if err != nil {
		return dbp, log.Error(dbp.ctx, 500, err, "Error connecting to dbp database")
	}
	err = dbp.conn.Ping()
	if err != nil {
		return dbp, log.Error(dbp.ctx, 500, err, "Connection to dbp database ping failed")
	}
	return dbp, nil
}

func GetDBPMySqlDSN() string {
	// Format: username:password@tcp(hostname:port)/database_name
	var result string
	username := os.Getenv("DBP_MYSQL_USERNAME")
	password := os.Getenv("DBP_MYSQL_PASSWORD")
	host := os.Getenv("DBP_MYSQL_HOST")
	port := os.Getenv("DBP_MYSQL_PORT")
	database := os.Getenv("DBP_MYSQL_DATABASE")
	result = username + ":" + password + "@tcp(" + host + ":" + port + ")/" + database
	return result
}

func (d *DBPAdapter) Close() {
	_ = d.conn.Close()
}

func (d *DBPAdapter) DeleteTimestamps(filesetId string) *log.Status {
	query := `DELETE FROM bible_file_timestamps WHERE bible_file_id IN
		(SELECT bf.id FROM bible_files bf, bible_filesets bs"
		WHERE bf.hash_id = bs.hash_id AND bs.id = ?)`
	_, err := d.conn.Exec(query, filesetId)
	if err != nil {
		return log.Error(d.ctx, 500, err, query)
	}
	return nil
}

func (d *DBPAdapter) SelectHashId(filesetId string) (string, *log.Status) {
	var result string
	query := `SELECT hash_id FROM bible_filesets WHERE asset_id = 'dbp-prod'
		AND set_type_code IN ('audio', 'audio_drama') AND id = ?")`
	rows, err := d.conn.Query(query, filesetId)
	if err != nil {
		return result, log.Error(d.ctx, 500, err, query)
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&result)
		if err != nil {
			return result, log.Error(d.ctx, 500, err, query)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.ctx, err, query)
	}
	return result, nil
}

func (d *DBPAdapter) UpdateFilesetTimingEstTag(hashId string, timingEstErr string) *log.Status {
	query := `REPLACE INTO bible_fileset_tags (hash_id, name, description, admin_only, iso, language_id)
		VALUES (?, 'timing_est_err', ?, 0, 'eng', 6414)")`
	_, err := d.conn.Exec(query, hashId, timingEstErr)
	if err != nil {
		return log.Error(d.ctx, 500, err, query)
	}
	return nil
}

func (d *DBPAdapter) SelectFileId(filesetId string, bookId string, chapterNum int) (int64, *log.Status) {
	var result int64
	query := `SELECT distinct bf.id FROM bible_files bf, bible_filesets bs
			WHERE bf.hash_id =  bs.hash_id
			AND bs.id = ? 
			AND bf.book_id = ?
			AND bf.chapter_start = ?`
	rows, err := d.conn.Query(query, filesetId, bookId, chapterNum)
	if err != nil {
		return result, log.Error(d.ctx, 500, err, query)
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&result)
		if err != nil {
			return result, log.Error(d.ctx, 500, err, query)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.ctx, err, query)
	}
	return result, nil
}

func (d *DBPAdapter) SelectTimestamps(filesetId string, bookId string, chapterNum int) ([]Segment, *log.Status) {
	var result []Segment
	query := `SELECT id, verse_start, timestamp FROM bible_file_timestamps WHERE bible_file_id IN
		(SELECT id FROM bible_files WHERE book_id=? AND chapter_start=? AND hash_id IN 
		(SELECT hash_id FROM bible_filesets WHERE id=?)) ORDER BY verse_sequence`
	rows, err := d.conn.Query(query, bookId, chapterNum, filesetId)
	if err != nil {
		return result, log.Error(d.ctx, 500, err, query)
	}
	defer rows.Close()
	for rows.Next() {
		var rec Segment
		err = rows.Scan(&rec.TimestampId, &rec.VerseStr, &rec.Timestamp)
		if err != nil {
			return result, log.Error(d.ctx, 500, err, query)
		}
		result = append(result, rec)
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.ctx, err, query)
	}
	return result, nil
}

func (d *DBPAdapter) UpdateTimestamps(bibleFileId int64, timestamps []db.Audio) *log.Status {
	query := `INSERT INTO bible_file_timestamps (bible_file_id, verse_start, verse_end,
		timestamp, timestamp_end, verse_sequence) VALUES (?,?,?,?,?,?)`
	tx, err := d.conn.Begin()
	if err != nil {
		return log.Error(d.ctx, 500, err, query)
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		return log.Error(d.ctx, 500, err, query)
	}
	defer stmt.Close()
	for _, rec := range timestamps {
		_, err = stmt.Exec(bibleFileId, rec.VerseStr, rec.VerseEnd, rec.BeginTS, rec.EndTS, rec.VerseSeq)
		if err != nil {
			return log.Error(d.ctx, 500, err, `Error while inserting dbp timestamp.`)
		}
		//records[i].ScriptId, err = qry.LastInsertId()
		//if err != nil {
		//	status = log.Error(d.Ctx, 500, err, `Error getting lastInsertId, while inserting Audio Verse.`)
		//	return records, status
		//}
	}
	err = tx.Commit()
	if err != nil {
		return log.Error(d.ctx, 500, err, query)
	}
	return nil
}

func (d *DBPAdapter) InsertBandwidth(bibleFileId int64, audioFile string, bandwidth string,
	codec string) (int64, *log.Status) {
	var lastInsertId int64
	query := `INSERT INTO bible_file_stream_bandwidths (bible_file_id, file_name, bandwidth, codec, stream)"
		VALUES (?, ?, ?, 'avc1.4d001f,mp4a.40.2', 1)`
	result, err := d.conn.Exec(query, bibleFileId, audioFile, bandwidth, codec)
	if err != nil {
		return lastInsertId, log.Error(d.ctx, 500, err, query)
	}
	lastInsertId, err = result.LastInsertId()
	if err != nil {
		return lastInsertId, log.Error(d.ctx, 500, err, query)
	}
	return lastInsertId, nil
}

func (d *DBPAdapter) InsertSegments(bandwidthId int64, segments []Segment) *log.Status {
	query := `INSERT INTO bible_file_stream_bytes (stream_bandwidth_id, timestamp_id, runtime, offset, bytes)"
	+ " VALUES (%s, %s, %s, %s, %s)`
	tx, err := d.conn.Begin()
	if err != nil {
		return log.Error(d.ctx, 500, err, query)
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		return log.Error(d.ctx, 500, err, query)
	}
	defer stmt.Close()
	for _, rec := range segments {
		_, err = stmt.Exec(bandwidthId, rec.TimestampId, rec.Duration, rec.Position, rec.NumBytes)
		if err != nil {
			return log.Error(d.ctx, 500, err, `Error while inserting dbp timestamp.`)
		}
	}
	err = tx.Commit()
	if err != nil {
		return log.Error(d.ctx, 500, err, query)
	}
	return nil
}
