package train

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	req "github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	"testing"
)

func TestNewWav2Vec2Trainer(t *testing.T) {
	ctx := context.Background()
	database, status := db.NewerDBAdapter(ctx, false, "GaryNTest", "N2KEUWB4")
	if status != nil {
		t.Fatal(status)
	}
	params := req.Wav2Vec2{BatchMB: 4, NumEpochs: 1, LearningRate: 1e-3, WarmupPct: 12, GradNormMax: 0.4}
	trainer := NewWav2Vec2Trainer(ctx, database, "keu", params)
	var f1 input.InputFile
	f1.Directory = "/Users/gary/FCBH2024/download/N2KEUWB4/N2KEUWBT"
	f1.Filename = "N2_KEU_WBT_069_JHN_001_VOX.mp3"
	f1.FileExt = ".mp3"
	f1.Testament = "nt"
	f1.BookId = "JHN"
	f1.BookSeq = "44"
	f1.Chapter = 1
	var files []input.InputFile
	files = append(files, f1)
	err := trainer.Train(files)
	if err != nil {
		t.Fatal(err)
	}
}
