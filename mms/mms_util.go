package mms

import (
	"bufio"
	"context"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/lang_tree/search"
	"io"
	"os/exec"
	"strings"
)

// Check that language is supported by mms_asr, and return alternate if it is not
func checkLanguage(ctx context.Context, lang string, sttLang string, aiTool string) (string, *log.Status) {
	var result string
	if sttLang != `` {
		result = sttLang
	} else {
		var tree = search.NewLanguageTree(ctx)
		err := tree.Load()
		if err != nil {
			return result, log.Error(ctx, 500, err, `Error loading language`)
		}
		langs, distance, err2 := tree.Search(strings.ToLower(lang), aiTool)
		if err2 != nil {
			return result, log.Error(ctx, 500, err2, `Error Searching for language`)
		}
		if len(langs) > 0 {
			result = langs[0]
			log.Info(ctx, `Using language`, result, "distance:", distance)
		} else {
			return result, log.ErrorNoErr(ctx, 400, `No compatible language code was found for`, lang)
		}
	}
	return result, nil
}

// deprecated - use utility.StdioExec
// callStdIOScript will exec the python script, and setup pipes on stdin and stdout
func callStdIOScript(ctx context.Context, command string, arg ...string) (*bufio.Writer, *bufio.Reader, *log.Status) {
	var writer *bufio.Writer
	var reader *bufio.Reader
	var status *log.Status
	cmd := exec.Command(command, arg...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		status = log.Error(ctx, 500, err, `Unable to open stdin for reading`)
		return writer, reader, status
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		status = log.Error(ctx, 500, err, `Unable to open stdout for writing`)
		return writer, reader, status
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		status = log.Error(ctx, 500, err, `Unable to open stderr for writing`)
		return writer, reader, status
	}
	err = cmd.Start()
	if err != nil {
		status = log.Error(ctx, 500, err, `Unable to start writing`)
		return writer, reader, status
	}
	handleStderr(ctx, stderr)
	writer = bufio.NewWriterSize(stdin, 4096)
	reader = bufio.NewReaderSize(stdout, 4096)
	return writer, reader, status
}

func handleStderr(ctx context.Context, stderr io.ReadCloser) {
	go func() {
		stderrReader := bufio.NewReader(stderr)
		for {
			line, err := stderrReader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					_ = log.Error(ctx, 500, err, "Error reading stderr")
				}
				return
			}
			log.Warn(ctx, "Stderr: ", line)
		}
	}()
}
