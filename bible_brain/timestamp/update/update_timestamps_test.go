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
			Timestamps: []string{"ENGESVN1DA", "ENGESVN1DA-opus16", "ENGESVN1SA", "ENGESVN2DA",
				"ENGESVN2DA-opus16", "ENGESVN2SA"},
		},
	}
	conn := db.NewDBAdapter(ctx, "test_data/ENGESVN1DA_TS.db")
	update := NewUpdateTimestamps(ctx, req, conn)
	status := update.Process()
	if status != nil {
		t.Fatal(status)
	}
}
