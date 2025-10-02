package stdio_exec

import (
	"bufio"
	"context"
	"encoding/binary"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"io"
	"os/exec"
	"strings"
)

type StdioExec struct {
	ctx     context.Context
	command string
	args    []string
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	writer  *bufio.Writer
	reader  *bufio.Reader
}

func NewStdioExec(ctx context.Context, command string, args ...string) (StdioExec, *log.Status) {
	var stdio StdioExec
	stdio.ctx = ctx
	stdio.command = command
	stdio.args = args
	var err error
	cmd := exec.Command(command, args...)
	stdio.stdin, err = cmd.StdinPipe()
	if err != nil {
		return stdio, log.Error(ctx, 500, err, `Unable to open stdin for reading`)
	}
	stdio.stdout, err = cmd.StdoutPipe()
	if err != nil {
		return stdio, log.Error(ctx, 500, err, `Unable to open stdout for writing`)
	}
	stdio.stderr, err = cmd.StderrPipe()
	if err != nil {
		return stdio, log.Error(ctx, 500, err, `Unable to open stderr for writing`)
	}
	err = cmd.Start()
	if err != nil {
		return stdio, log.Error(ctx, 500, err, `Unable to start writing`)
	}
	handleStderr(ctx, stdio.stderr)
	stdio.writer = bufio.NewWriterSize(stdio.stdin, 4096)
	stdio.reader = bufio.NewReaderSize(stdio.stdout, 4096)
	return stdio, nil
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

func (s *StdioExec) Process(input string) (string, *log.Status) {
	var result string
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
	result = strings.TrimRight(result, "\n")
	return result, nil
}

func (s *StdioExec) ProcessBytes(input []byte) (string, *log.Status) {
	var result string
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
	result = strings.TrimRight(result, "\n")
	return result, nil
}

func (s *StdioExec) Close() {
	_ = s.stdin.Close()
	_ = s.stdout.Close()
}
