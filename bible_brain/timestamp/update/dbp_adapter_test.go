package update

import (
	"context"
	"fmt"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"testing"
)

func TestNewDBPAdapter(t *testing.T) {
	ctx := context.Background()
	_, status := NewDBPAdapter(ctx)
	if status != nil {
		t.Fatal(status)
	}

}

func TestNewDBPAdapter2(t *testing.T) {
	ctx := context.Background()
	dbp, status := NewDBPAdapter(ctx)
	if status != nil {
		t.Fatal(status)
	}
	var count int
	rows, err := dbp.conn.Query("SELECT count(*) from bible_files")
	if err != nil {
		status = log.Error(ctx, 500, err, "err")
		t.Fatal(status)
	}
	//defer d.closeDef(rows, "SelectScalarInt stmt")
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			status = log.Error(ctx, 500, err, "")
			t.Fatal(status)
		}
	}
	err = rows.Err()
	if err != nil {
		t.Fatal(status)
	}
	fmt.Println("count", count)
}
