package adapter

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	req "github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/ffmpeg"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/stdio_exec"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type TrainAdapter struct {
	ctx     context.Context
	conn    db.DBAdapter
	langISO string
	args    req.MMSAdapter
}

func NewTrainAdapter(ctx context.Context, conn db.DBAdapter, langISO string, train req.MMSAdapter) TrainAdapter {
	var t TrainAdapter
	t.ctx = ctx
	t.conn = conn
	ident, status := t.conn.SelectIdent()
	fmt.Println("Status: ", status)
	fmt.Println("Ident: ", ident)
	scripts, status := t.conn.SelectScripts()
	fmt.Println("Status: ", status, "Len Scripts: ", len(scripts))
	t.langISO = langISO
	t.args = train
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
	pythonScript := filepath.Join(os.Getenv("GOPROJ"), "mms/adapter/trainer.py")
	status := stdio_exec.RunScriptWithLogging(t.ctx, pythonPath, pythonScript,
		t.langISO,
		t.conn.DatabasePath,
		strings.Replace(tempDir, ` `, `\ `, -1),
		strconv.Itoa(t.args.BatchMB),
		strconv.Itoa(t.args.NumEpochs),
		strconv.FormatFloat(t.args.LearningRate, 'e', -1, 64),
		strconv.FormatFloat(t.args.WarmupPct, 'f', -1, 64),
		strconv.FormatFloat(t.args.GradNormMax, 'f', -1, 64))
	return status
}
