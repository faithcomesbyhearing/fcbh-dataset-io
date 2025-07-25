package update

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"testing"
)

func TestUpdateValidation(t *testing.T) {
	ctx := context.Background()
	dbpConn, status := NewDBPAdapter(ctx)
	//dbpConn := db.NewDBAdapter(ctx, "test_data/ENGESVN1DA_TS.db")
	if status != nil {
		t.Fatal(status)
	}
	defer dbpConn.Close()
	conn := db.NewDBAdapter(ctx, "test_data/ENGESVN1DA_TS.db")
	defer conn.Close()
	status = updateValidation(ctx, conn, dbpConn)
	if status != nil {
		t.Fatal(status)
	}
}
