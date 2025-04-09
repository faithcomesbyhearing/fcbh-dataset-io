package update

import (
	"context"
	"database/sql"
	"errors"
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
	mysqlDSN := os.Getenv("DBP_MYSQL_DSN")
	mysqlDSN += "?clientFoundRows=true" // makes affected row count number matched, not number updated
	dbp.conn, err = sql.Open("mysql", mysqlDSN)
	if err != nil {
		return dbp, log.Error(dbp.ctx, 500, err, "Error connecting to dbp database")
	}
	err = dbp.conn.Ping()
	if err != nil {
		return dbp, log.Error(dbp.ctx, 500, err, "Connection to dbp database ping failed")
	}
	return dbp, nil
}

func (d *DBPAdapter) Close() {
	_ = d.conn.Close()
}

func (d *DBPAdapter) SelectHashId(filesetId string) (string, *log.Status) {
	var result string
	query := `SELECT hash_id FROM bible_filesets WHERE asset_id = 'dbp-prod'
		AND set_type_code IN ('audio', 'audio_drama', 'audio_stream', 'audio_drama_stream') 
		AND id = ?`
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

func (d *DBPAdapter) SelectFileId(hashId string, bookId string, chapterNum int) (int64, string, *log.Status) {
	var result int64
	var filename string
	query := `SELECT distinct id, file_name FROM bible_files WHERE hash_id = ? AND book_id = ? and chapter_start = ?`
	rows, err := d.conn.Query(query, hashId, bookId, chapterNum)
	if err != nil {
		return result, filename, log.Error(d.ctx, 500, err, query)
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&result, &filename)
		if err != nil {
			return result, filename, log.Error(d.ctx, 500, err, query)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.ctx, err, query)
	}
	return result, filename, nil
}

func (d *DBPAdapter) SelectTimestamps(fileId int64) ([]Timestamp, *log.Status) {
	var result []Timestamp
	query := `SELECT id, verse_start, verse_end, verse_sequence, timestamp, timestamp_end 
		FROM bible_file_timestamps WHERE bible_file_id = ? 
		ORDER BY verse_sequence, id`
	rows, err := d.conn.Query(query, fileId)
	if err != nil {
		return result, log.Error(d.ctx, 500, err, query)
	}
	defer rows.Close()
	for rows.Next() {
		var tmpEndTS sql.NullFloat64
		var rec Timestamp
		err = rows.Scan(&rec.TimestampId, &rec.VerseStr, &rec.VerseEnd,
			&rec.VerseSeq, &rec.BeginTS, &tmpEndTS)
		if err != nil {
			return result, log.Error(d.ctx, 500, err, query)
		}
		// How should I handle verse end
		if tmpEndTS.Valid {
			rec.EndTS = tmpEndTS.Float64
		}
		result = append(result, rec)
	}
	err = rows.Err()
	if err != nil {
		log.Warn(d.ctx, err, query)
	}
	return result, nil
}

func (d *DBPAdapter) UpdateTimestamps(timestamps []Timestamp) (int, *log.Status) {
	var rowCount int
	var mustUpdate int
	for _, rec := range timestamps {
		if rec.TimestampId > 0 {
			mustUpdate++
		}
	}
	if mustUpdate > 0 {
		query := `UPDATE bible_file_timestamps SET timestamp = ?, timestamp_end = ? 
			WHERE id = ?`
		tx, err := d.conn.Begin()
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		stmt, err := tx.Prepare(query)
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		defer stmt.Close()
		var result sql.Result
		var count int64
		for _, rec := range timestamps {
			if rec.TimestampId > 0 {
				result, err = stmt.Exec(rec.BeginTS, rec.EndTS, rec.TimestampId)
				if err != nil {
					return rowCount, log.Error(d.ctx, 500, err, query)
				}
				count, err = result.RowsAffected()
				if err != nil {
					return rowCount, log.Error(d.ctx, 500, err, query)
				}
				rowCount += int(count)
			}
		}
		err = tx.Commit()
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		if rowCount != mustUpdate {
			return rowCount, log.ErrorNoErr(d.ctx, 500, "Row count expected:",
				mustUpdate, "Actual Count:", rowCount, query)
		}
	}
	return rowCount, nil
}

func (d *DBPAdapter) InsertTimestamps(bibleFileId int64, timestamps []Timestamp) ([]Timestamp, int, *log.Status) {
	var rowCount int
	var mustInsert int
	for _, rec := range timestamps {
		if rec.TimestampId == 0 {
			mustInsert++
		}
	}
	if mustInsert > 0 {
		query := `INSERT INTO bible_file_timestamps (bible_file_id, verse_start, verse_end,
		timestamp, timestamp_end, verse_sequence) VALUES (?,?,?,?,?,?)`
		tx, err := d.conn.Begin()
		if err != nil {
			return timestamps, rowCount, log.Error(d.ctx, 500, err, query)
		}
		stmt, err := tx.Prepare(query)
		if err != nil {
			return timestamps, rowCount, log.Error(d.ctx, 500, err, query)
		}
		defer stmt.Close()
		var result sql.Result
		var count int64
		for i, rec := range timestamps {
			if rec.TimestampId == 0 {
				result, err = stmt.Exec(bibleFileId, rec.VerseStr, rec.VerseEnd, rec.BeginTS, rec.EndTS, rec.VerseSeq)
				if err != nil {
					return timestamps, rowCount, log.Error(d.ctx, 500, err, `Error while inserting dbp timestamp.`)
				}
				timestamps[i].TimestampId, err = result.LastInsertId()
				if err != nil {
					return timestamps, rowCount, log.Error(d.ctx, 500, err, `Error getting lastInsertId, while inserting Timestamps.`)
				}
				count, err = result.RowsAffected()
				if err != nil {
					return timestamps, rowCount, log.Error(d.ctx, 500, err, query)
				}
				rowCount += int(count)
			}
		}
		err = tx.Commit()
		if err != nil {
			return timestamps, rowCount, log.Error(d.ctx, 500, err, query)
		}
		if rowCount != mustInsert {
			return timestamps, rowCount, log.ErrorNoErr(d.ctx, 500,
				"Row count expected:", mustInsert, "Actual Count:", rowCount, query)
		}
	}
	return timestamps, rowCount, nil
}

func (d *DBPAdapter) UpdateSegments(segments []Timestamp) (int, *log.Status) {
	var rowCount int
	query := `UPDATE bible_file_stream_bytes SET runtime = ?, offset = ?, bytes = ?
		WHERE timestamp_id = ?`
	tx, err := d.conn.Begin()
	if err != nil {
		return rowCount, log.Error(d.ctx, 500, err, query)
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		return rowCount, log.Error(d.ctx, 500, err, query)
	}
	defer stmt.Close()
	var result sql.Result
	var count int64
	for _, rec := range segments {
		result, err = stmt.Exec(rec.Duration, rec.Position, rec.NumBytes, rec.TimestampId)
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, `Error while inserting dbp timestamp.`)
		}
		count, err = result.RowsAffected()
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		rowCount += int(count)
	}
	err = tx.Commit()
	if err != nil {
		return rowCount, log.Error(d.ctx, 500, err, query)
	}
	return rowCount, nil
}

func (d *DBPAdapter) UpdateFilesetTimingEstTag(hashId string, timingEstErr string) (int, *log.Status) {
	var rowCount int
	query := `SELECT description FROM bible_fileset_tags WHERE hash_id = ? AND name = 'timing_est_err'`
	var currEstErr string
	err := d.conn.QueryRow(query, hashId).Scan(&currEstErr)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, log.Error(d.ctx, 500, err, query)
	}
	var result sql.Result
	var count int64
	if errors.Is(err, sql.ErrNoRows) {
		query = `INSERT INTO bible_fileset_tags (hash_id, name, description, admin_only, iso, language_id)
		VALUES (?, 'timing_est_err', ?, 0, 'eng', 6414)`
		result, err = d.conn.Exec(query, hashId, timingEstErr)
		if err != nil {
			return 0, log.Error(d.ctx, 500, err, query)
		}
		count, err = result.RowsAffected()
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		rowCount = int(count)
	} else if currEstErr != timingEstErr {
		query = `UPDATE bible_fileset_tags SET description = ? WHERE hash_id = ? AND name = 'timing_est_err'`
		result, err = d.conn.Exec(query, timingEstErr, hashId)
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		count, err = result.RowsAffected()
		if err != nil {
			return rowCount, log.Error(d.ctx, 500, err, query)
		}
		rowCount += int(count)
		if rowCount != 1 {
			return rowCount, log.ErrorNoErr(d.ctx, 500,
				"Row count expected:", 1, "Actual Count:", rowCount, query)
		}
	}
	return rowCount, nil
}
