package cleanup

import (
	"context"
	"os"
	"path/filepath"
	"time"

	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

func CleanupDownloadDirectory(ctx context.Context) {
	downloadDir := os.Getenv("FCBH_DATASET_FILES")
	maxAge := 3 * 24 * time.Hour // 3 day
	_ = CleanupDirectory(ctx, downloadDir, maxAge)
	tmpDir := os.Getenv("FCBH_DATASET_TMP")
	_ = CleanupDirectory(ctx, tmpDir, maxAge)
}

func CleanupDirectory(ctx context.Context, directory string, maxAge time.Duration) *log.Status {
	now := time.Now()
	count := 0
	entries, err := os.ReadDir(directory)
	if err != nil {
		return log.Error(ctx, 500, err, "Error reading directory", directory)
	}
	for _, entry := range entries {
		dirPath := filepath.Join(directory, entry.Name())
		var info os.FileInfo
		info, err = os.Stat(dirPath)
		if err != nil {
			log.Warn(ctx, "Unable to stat directory", dirPath, err)
			continue
		}
		if now.Sub(info.ModTime()) > maxAge {
			err = os.RemoveAll(dirPath)
			if err != nil {
				log.Warn(ctx, "Unable to remove directory", dirPath, err)
				continue
			}
			count++
		}
	}
	log.Info(ctx, "Removed from directory", directory, count)
	return nil
}
