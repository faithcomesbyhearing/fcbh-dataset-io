package adapter

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	req "github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/ffmpeg"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/stdio_exec"
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
	t.langISO = langISO
	t.args = train
	return t
}

func (t *TrainAdapter) HasModel() bool {
	filename := "adapter." + t.langISO + ".safetensors"
	model := filepath.Join(os.Getenv("FCBH_DATASET_DB"), "mms_adapters", t.langISO, filename)
	fileInfo, err := os.Stat(model)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		log.Warn(t.ctx, err, "Failed to read model file")
		return false
	}
	return fileInfo.Size() > 1000000 // must be GT 1Meg
}

func (t *TrainAdapter) Train(files []input.InputFile) *log.Status {
	if len(files) == 0 {
		return nil
	}
	tempDir, err := os.MkdirTemp(os.Getenv(`FCBH_DATASET_TMP`), "mms_adapter_")
	if err != nil {
		return log.Error(t.ctx, 500, err, `Error creating temp dir`)
	}
	defer os.RemoveAll(tempDir)
	for _, file := range files {
		_, status := ffmpeg.ConvertMp3ToWav(t.ctx, tempDir, file.FilePath())
		if status != nil {
			return status
		}
	}
	scriptsNum, status := t.conn.CountScriptRows()
	if status != nil {
		return status
	}
	threshold := int(math.Ceil(float64(scriptsNum) * 0.20))
	status = SilencePruner(t.ctx, threshold, t.conn)
	if status != nil {
		return status
	}
	pythonPath := os.Getenv(`FCBH_MMS_ADAPTER_PYTHON`)
	pythonScript := filepath.Join(os.Getenv("GOPROJ"), "mms/adapter/trainer.py")
	status = stdio_exec.RunScriptWithLogging(t.ctx, pythonPath, pythonScript,
		t.langISO,
		t.conn.DatabasePath,
		`'`+tempDir+`'`,
		strconv.Itoa(t.args.BatchMB),
		strconv.Itoa(t.args.NumEpochs),
		strconv.FormatFloat(t.args.LearningRate, 'e', -1, 64),
		strconv.FormatFloat(t.args.WarmupPct, 'f', -1, 64),
		strconv.FormatFloat(t.args.GradNormMax, 'f', -1, 64))
	return status
}
