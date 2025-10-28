package stdio_exec

import (
	"bufio"
	"context"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"os/exec"
	"strings"
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
	cmd := exec.CommandContext(ctx, python, newArgs...)
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
	var pythonErr *log.Status
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) > 0 {
				log.Info(ctx, "PY:", line)
			}
		}
	}()
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) > 0 {
				status := log.ExecError(ctx, 500, line)
				if status != nil {
					pythonErr = status
				}
			}
		}
		err = scanner.Err()
		if err != nil {
			_ = log.Error(ctx, 500, err, "Error reading stderr")
		}
	}()
	wg.Wait() // Wait for goroutines to finish reading any remaining output
	err = cmd.Wait()
	if err != nil {
		// Log, but discard so that error caught in python is returned
		_ = log.Error(ctx, 500, err, `Module failed`, cmd.String())
	}
	return pythonErr
}
