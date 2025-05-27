package input

import (
	"context"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

func FileInput(ctx context.Context, path string) ([]InputFile, *log.Status) {
	var files []InputFile
	var status *log.Status
	files, status = Glob(ctx, path)
	return files, status
}
