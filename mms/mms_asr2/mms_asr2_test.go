package mms_asr

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/match/diff"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/stdio_exec"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/uroman"
	"os"
	"path"
	"strconv"
	"testing"
)

func TestMMSASR2_ProcessFiles(t *testing.T) {
	ctx := context.Background()
	log.SetOutput("stderr")
	user := request.GetTestUser()
	conn, status := db.NewerDBAdapter(ctx, false, user, "N2MZJSIM") // is not used
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

func TestMMSASR2_ParseResult(t *testing.T) {
	ctx := context.Background()
	log.SetOutput("stderr")
	user := request.GetTestUser()
	conn, status := db.NewerDBAdapter(ctx, false, user, "N2MZJSIM")
	if status != nil {
		t.Fatal(status)
	}
	asr := NewMMSASR2(ctx, conn, "mzj", "", false)
	var file input.InputFile
	file.BookId = "MAT"
	file.Chapter = 12
	file.MediaId = "N2MZJSIM"
	asr.uroman, status = stdio_exec.NewStdioExec(asr.ctx, os.Getenv(`FCBH_MMS_FA_PYTHON`), uroman.ScriptPath(), "-l", asr.lang)
	if status != nil {
		t.Fatal(status)
	}
	defer func() {
		status = asr.uroman.Close()
	}()
	response := readResultFile(file)
	status = asr.parseResult(file, response)

	remove := request.CompareChoice{Remove: true}
	nfc := request.DiacriticalChoice{NormalizeNFC: true}
	compare := diff.NewCompare(ctx, user, "", conn, "mzj", request.Testament{NT: true},
		request.CompareSettings{LowerCase: true, RemovePromptChars: true, RemovePunctuation: true,
			DoubleQuotes: remove, Apostrophe: remove, Hyphen: remove, DiacriticalMarks: nfc})
	pairs, fileMap, lang, status2 := compare.CompareASR2()
	if status2 != nil {
		t.Fatal(status2)
	}
	fmt.Println("num pairs", len(pairs))
	report := diff.NewHTMLWriter(ctx, "datasetName1")
	filename, status := report.WriteReport("baseDataetName", pairs, lang, fileMap, request.SpeechToText{MMS: true})
	if status != nil {
		t.Fatal(status)
	}
	fmt.Println("htmlFilename", filename)
}

func readResultFile(file input.InputFile) string {
	content, err := os.ReadFile(file.MediaId + "_" + file.BookId + strconv.Itoa(file.Chapter))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return string(content)
}
