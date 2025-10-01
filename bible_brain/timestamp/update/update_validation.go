package update

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

func updateValidation(ctx context.Context, apAdapter db.DBAdapter, dbpAdapter DBPAdapter) *log.Status {
	var apStats, dbpStats stats
	var status *log.Status
	apStats, status = getStats(ctx, apAdapter.DB)
	if status != nil {
		return status
	}
	dbpStats, status = getStats(ctx, dbpAdapter.conn)
	if status != nil {
		return status
	}
	var msgs []string
	if apStats.beginTSCount != dbpStats.beginTSCount {
		msg := fmt.Sprintf("Begin TS count mismatch, ap: %d dbp: %d", apStats.beginTSCount, dbpStats.beginTSCount)
		msgs = append(msgs, msg)
	}
	if apStats.endTSCount != dbpStats.endTSCount {
		msg := fmt.Sprintf("End TS ecount mismatch, ap: %d dbp: %d", apStats.endTSCount, dbpStats.endTSCount)
		msgs = append(msgs, msg)
	}
	if apStats.beginTSSum != dbpStats.beginTSSum {
		msg := fmt.Sprintf("Begin TS sum mismatch, ap: %f dbp: %f", apStats.beginTSSum, dbpStats.beginTSSum)
		msgs = append(msgs, msg)
	}
	if apStats.endTSSum != dbpStats.endTSSum {
		msg := fmt.Sprintf("End TS sum mismatch, ap: %f dbp: %f", apStats.beginTSSum, dbpStats.beginTSSum)
		msgs = append(msgs, msg)
	}
	if len(msgs) > 0 {
		return log.ErrorNoErr(ctx, 400, strings.Join(msgs, "\n"))
	} else {
		log.Info(ctx, "BeginTS", apStats.beginTSCount, apStats.beginTSSum,
			"EndTS", apStats.endTSCount, apStats.endTSSum)
	}
	return nil
}

type stats struct {
	beginTSCount int64
	endTSCount   int64
	beginTSSum   float64
	endTSSum     float64
}

func getStats(ctx context.Context, conn *sql.DB) (stats, *log.Status) {
	// Try to determine if this is SQLite or MySQL by checking if scripts table exists
	var query string
	var tableName string

	// Check if scripts table exists (SQLite database)
	checkQuery := `SELECT name FROM sqlite_master WHERE type='table' AND name='scripts'`
	var tableExists bool
	err := conn.QueryRow(checkQuery).Scan(&tableExists)
	if err == nil && tableExists {
		// This is SQLite with scripts table
		tableName = "scripts"
		query = `SELECT count(script_begin_ts), count(script_end_ts), sum(script_begin_ts), sum(script_end_ts)
				FROM scripts`
	} else {
		// This is MySQL with bible_file_timestamps table
		tableName = "bible_file_timestamps"
		query = `SELECT count(begin_ts), count(end_ts), sum(begin_ts), sum(end_ts)
				FROM bible_file_timestamps`
	}

	row := conn.QueryRow(query)
	var st stats
	err = row.Scan(&st.beginTSCount, &st.endTSCount, &st.beginTSSum, &st.endTSSum)
	if err != nil {
		return st, log.Error(ctx, 500, err, `Error in getStats querying `+tableName)
	}
	return st, nil
}
