package mms_asr

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"os"
	"path"
	"testing"
)

func TestMMSASR2_ProcessFiles(t *testing.T) {
	ctx := context.Background()
	log.SetOutput("stderr")
	//conn := db.NewDBAdapter(ctx, ":memory:")
	user := request.GetTestUser()
	conn, status := db.NewerDBAdapter(ctx, true, user, "N2MZJSIM")
	asr := NewMMSASR2(ctx, conn, "mzj", "", false)
	var files []input.InputFile
	var file input.InputFile
	file.BookId = "MAT"
	file.Chapter = 12
	file.MediaId = "N2MZJSIM"
	file.Directory = path.Join(os.Getenv("FCBH_DATASET_FILES"), "N2MZJSIM", "N2MZJSIM Chapter VOX")
	file.Filename = "N2_MZJ_SIM_012_MAT_012_VOX.mp3"
	fmt.Println("audio file: ", file.FilePath())
	files = append(files, file)
	status = asr.ProcessFiles(files)
	if status != nil {
		t.Fatal(status)
	}
}
