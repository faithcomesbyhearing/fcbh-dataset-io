package stdio_exec

import (
	"bufio"
	"context"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"os/exec"
	"sync"
)

/**
This func will run a long running process, such as a python training program,
and capture the stdout and stderr output, It will log the stdout lines as INFO,
and the STDERR lines as WARN.
*/

func RunScriptWithLogging(ctx context.Context, python string, args ...string) *log.Status {
	var newArgs []string
	newArgs = append(newArgs, "-u")
	newArgs = append(newArgs, args...)
	cmd := exec.Command(python, newArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return log.Error(ctx, 500, err, `Unable to open stdout for writing`, cmd.String())
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return log.Error(ctx, 500, err, `Unable to open stderr for writing`, cmd.String())
	}
	err = cmd.Start()
	if err != nil {
		return log.Error(ctx, 500, err, `Unable to execute command`, cmd.String())
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log.Info(ctx, "PY:", scanner.Text())
		}
	}()
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Warn(ctx, "PY:", scanner.Text())
		}
	}()
	err = cmd.Wait() // Wait for process to complete
	if err != nil {
		return log.Error(ctx, 500, err, `Error occurred in final wait of cmd`, cmd.String())
	}
	wg.Wait() // Wait for goroutines to finish reading any remaining output
	return nil
}

/**
// This code splits on CR and LF, but the idea could be used to split on larger pieces.
scanner := bufio.NewScanner(stdout)
scanner.Split(splitOnNewlineOrCarriageReturn)

func splitOnNewlineOrCarriageReturn(data []byte, atEOF bool) (advance int, token []byte, err error) {
    if atEOF && len(data) == 0 {
        return 0, nil, nil
    }
    // Look for \n or \r
    if i := bytes.IndexAny(data, "\n\r"); i >= 0 {
        return i + 1, data[0:i], nil
    }
    if atEOF {
        return len(data), data, nil
    }
    return 0, nil, nil
}
*/
