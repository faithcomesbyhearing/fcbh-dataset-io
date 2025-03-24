package update

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"testing"
)

func TestUpdateTimestamps_Process(t *testing.T) {
	log.SetOutput("stdout")
	ctx := context.Background()
	req := request.Request{
		UpdateDBP: request.UpdateDBP{
			Timestamps: []string{"ENGESVN1DA", "ENGESVN1SA", "ENGESVN2DA", "ENGESVN2SA"},
		},
	}
	conn := db.NewDBAdapter(ctx, "test_data/ENGESVN1DA_TS.db")
	update := NewUpdateTimestamps(ctx, req, conn)
	status := update.Process()
	if status != nil {
		t.Fatal(status)
	}
}

// The following have no timestamps:
// AAAMLTN1DA_TS.yaml
// ABIWBTN1DA_TS.yaml
// ABPWBTN1DA_TS.yaml
// ACCIBSN1DA_TS.yaml
func TestInsertTimestamps_Process(t *testing.T) {
	log.SetOutput("stdout")
	ctx := context.Background()
	req := request.Request{
		UpdateDBP: request.UpdateDBP{
			Timestamps: []string{"AAAMLTN1DA", "AAAMLTN1DA-opus16", "AAAMLTN2DA",
				"AAAMLTN2DA-opus16"},
		},
	}
	conn := db.NewDBAdapter(ctx, "test_data/ENGESVN1DA_TS.db")
	update := NewUpdateTimestamps(ctx, req, conn)
	status := update.Process()
	if status != nil {
		t.Fatal(status)
	}
}
