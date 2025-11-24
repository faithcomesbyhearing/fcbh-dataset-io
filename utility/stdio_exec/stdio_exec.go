package stdio_exec

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"os/exec"
	"strings"
	"sync"

	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

type StdioExec struct {
	ctx       context.Context
	command   string
	args      []string
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	writer    *bufio.Writer
	reader    *bufio.Reader
	stderrWg  sync.WaitGroup
	pythonErr *log.Status
	errMutex  sync.Mutex
}

func NewStdioExec(ctx context.Context, command string, args ...string) (*StdioExec, *log.Status) {
	var stdio StdioExec
	stdio.ctx = ctx
	stdio.command = command
	stdio.args = args
	var err error
	stdio.cmd = exec.CommandContext(ctx, command, args...)
	stdio.stdin, err = stdio.cmd.StdinPipe()
	if err != nil {
		return &stdio, log.Error(ctx, 500, err, `Unable to open stdin for reading`)
	}
	stdio.stdout, err = stdio.cmd.StdoutPipe()
	if err != nil {
		return &stdio, log.Error(ctx, 500, err, `Unable to open stdout for writing`)
	}
	stdio.stderr, err = stdio.cmd.StderrPipe()
	if err != nil {
		return &stdio, log.Error(ctx, 500, err, `Unable to open stderr for writing`)
	}
	err = stdio.cmd.Start()
	if err != nil {
		return &stdio, log.Error(ctx, 500, err, `Unable to start writing`)
	}
	stdio.handleStderr()
	stdio.writer = bufio.NewWriterSize(stdio.stdin, 4096)
	stdio.reader = bufio.NewReaderSize(stdio.stdout, 4096)
	return &stdio, nil
}

func (s *StdioExec) handleStderr() {
	s.stderrWg.Add(1)
	go func() {
		defer s.stderrWg.Done()
		scanner := bufio.NewScanner(s.stderr)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) > 0 {
				status := log.ExecError(s.ctx, 500, line)
				if status != nil {
					s.errMutex.Lock()
					s.pythonErr = status
					s.errMutex.Unlock()
				}
			}
		}
		err := scanner.Err()
		if err != nil {
			_ = log.Error(s.ctx, 500, err, "Error reading stderr")
		}
	}()
}

func (s *StdioExec) getPythonErr() *log.Status {
	s.errMutex.Lock()
	defer s.errMutex.Unlock()
	return s.pythonErr
}

func (s *StdioExec) Process(input string) (string, *log.Status) {
	var result string
	pyErr := s.getPythonErr()
	if pyErr != nil {
		return result, pyErr
	}
	_, err := s.writer.WriteString(input + "\n")
	if err != nil {
		return result, log.Error(s.ctx, 500, err, "Error writing to", s.command)
	}
	err = s.writer.Flush()
	if err != nil {
		return result, log.Error(s.ctx, 500, err, "Error flush to", s.command)
	}
	result, err = s.reader.ReadString('\n')
	if err != nil {
		return result, log.Error(s.ctx, 500, err, `Error reading response from`, s.command)
	}
	pyErr = s.getPythonErr()
	if pyErr != nil {
		return result, pyErr
	}
	result = strings.TrimRight(result, "\n")
	return result, nil
}

func (s *StdioExec) ProcessBytes(input []byte) (string, *log.Status) {
	var result string
	pyErr := s.getPythonErr()
	if pyErr != nil {
		return result, pyErr
	}
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(input)))
	_, err := s.writer.Write(lengthBuf)
	if err != nil {
		return result, log.Error(s.ctx, 500, err, "Error writing length to", s.command)
	}
	_, err = s.writer.Write(input)
	if err != nil {
		return result, log.Error(s.ctx, 500, err, "Error writing to", s.command)
	}
	err = s.writer.Flush()
	if err != nil {
		return result, log.Error(s.ctx, 500, err, "Error flush to", s.command)
	}
	result, err = s.reader.ReadString('\n')
	if err != nil {
		return result, log.Error(s.ctx, 500, err, `Error reading response from`, s.command)
	}
	pyErr = s.getPythonErr()
	if pyErr != nil {
		return result, pyErr
	}
	result = strings.TrimRight(result, "\n")
	return result, nil
}

func (s *StdioExec) Close() {
	if s.writer != nil {
		_ = s.writer.Flush()
	}
	if s.stdin != nil {
		_ = s.stdin.Close()
	}
	s.stderrWg.Wait()
	if s.cmd != nil && s.cmd.Process != nil {
		err := s.cmd.Wait()
		if err != nil {
			// Do not return error so that s.pythonErr is reported
			_ = log.Error(s.ctx, 500, err, `Module failed`, s.cmd.String())
		}
	}
	return
}
