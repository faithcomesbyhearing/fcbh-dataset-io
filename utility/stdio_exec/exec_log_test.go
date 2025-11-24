package stdio_exec

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

func TestRunScriptWithLogging(t *testing.T) {
	log.SetOutput("stdout")
	ctx := context.Background()
	pythonPath := os.Getenv(`FCBH_MMS_ADAPTER_PYTHON`)
	var execLogTest = `import os
import sys
sys.path.insert(0, os.path.abspath(os.path.join(os.environ['GOPROJ'], 'logger')))
from error_handler import setup_error_handler
setup_error_handler()
print("12", 2 * 6, file=sys.stdout, flush=True)
#sys.exit(10)
print("24", 4 * 6, file=sys.stderr, flush=True)
print("Inf", 5 / 0, flush=True)
print("36", 6 * 6, flush=True)
print("64", 8 * 8, file=sys.stderr, flush=True)
`
	pythonScript := filepath.Join(os.Getenv("FCBH_DATASET_TMP"), "exec_log_test.py")
	err := os.WriteFile(pythonScript, []byte(execLogTest), 0644)
	if err != nil {
		t.Fatal(err)
	}
	status := RunScriptWithLogging(ctx, pythonPath, pythonScript)
	fmt.Println("test result status", status)
}
