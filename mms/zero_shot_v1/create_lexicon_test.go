package zero_shot_v1

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"testing"
)

func TestCreateLexicon(t *testing.T) {
	ctx := context.Background()
	database, status := db.NewerDBAdapter(ctx, false, "GaryNTest", "N2CUL_MNT")
	if status != nil {
		t.Fatal(status)
	}
	defer database.Close()
	directory, status := createLexiconFile(ctx, database.DB)
	if status != nil {
		t.Fatal(status)
	}
	print(directory)
}
