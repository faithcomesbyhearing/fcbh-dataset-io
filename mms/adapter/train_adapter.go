package adapter

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/ffmpeg"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/stdio_exec"
	"os"
	"path/filepath"
	"strconv"
)

type TrainAdapter struct {
	ctx         context.Context
	conn        db.DBAdapter
	langISO     string
	batchSizeMB int
	epochs      int
	restart     string
}

func NewTrainAdapter(ctx context.Context, conn db.DBAdapter, langISO string, batchSize int, epochs int) TrainAdapter {
	var t TrainAdapter
	t.ctx = ctx
	t.conn = conn
	ident, status := t.conn.SelectIdent()
	fmt.Println("Status: ", status)
	fmt.Println("Ident: ", ident)
	scripts, status := t.conn.SelectScripts()
	fmt.Println("Status: ", status, "Len Scripts: ", len(scripts))
	t.langISO = langISO
	t.batchSizeMB = batchSize
	t.epochs = epochs
	return t
}

func (t *TrainAdapter) Train(files []input.InputFile) *log.Status {
	if len(files) == 0 {
		return nil
	}
	tempDir := files[0].Directory
	for _, file := range files {
		_, status := ffmpeg.ConvertMp3ToWav(t.ctx, tempDir, file.FilePath())
		if status != nil {
			return status
		}
	}
	pythonPath := os.Getenv(`FCBH_MMS_ADAPTER_PYTHON`)
	pythonScript := filepath.Join(os.Getenv("GOPROJ"), "mms/adapter/train_adapter.py")
	status := stdio_exec.RunScriptWithLogging(t.ctx, pythonPath, pythonScript,
		t.langISO,
		t.conn.DatabasePath,
		tempDir,
		strconv.Itoa(t.batchSizeMB),
		strconv.Itoa(t.epochs))
	return status
}
