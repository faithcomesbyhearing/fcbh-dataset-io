package input

import (
	"context"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"io"
	"os"
	"path/filepath"
)

/*
PostFiles contains the filenames of the files posted to the server using a mulipart form.
The key in the form should be "text" or "audio", and it ends in text or audio file set.
*/

type PostFiles struct {
	ctx   context.Context
	dir   string
	text  []InputFile
	audio []InputFile
}

func NewPostFiles(ctx context.Context) PostFiles {
	var p PostFiles
	p.ctx = ctx
	var err error
	p.dir, err = os.MkdirTemp(os.Getenv("FCBH_DATASET_TMP"), "post*")
	if err != nil {
		log.Warn(ctx, "Failed to create temp dir in NewPostFiles")
	}
	return p
}

func (p *PostFiles) ReadFile(ftype string, source io.Reader, filename string) *log.Status {
	var status *log.Status
	var file InputFile
	target, err := os.Create(filepath.Join(p.dir, filename))
	if err != nil {
		return log.Error(p.ctx, 500, err, "Failed to create temp directory for post")
	}
	defer target.Close()
	_, err = io.Copy(target, source)
	if err != nil {
		return log.Error(p.ctx, 500, err, "Failed to save audio file")
	}
	file.Filename = filename
	file.Directory = filepath.Dir(target.Name())
	if ftype == "text" {
		p.text = append(p.text, file)
	} else if ftype == "audio" {
		p.audio = append(p.audio, file)
	} else {
		status = log.Error(p.ctx, 500, err, "Invalid file type")
	}
	return status
}

func (p *PostFiles) PostInput(ftype string) ([]InputFile, *log.Status) {
	var files []InputFile
	if ftype == "text" {
		files = p.text
	} else if ftype == "audio" {
		files = p.audio
	}
	return files, nil
}

func (p *PostFiles) RemoveDir() {
	_ = os.RemoveAll(p.dir)
}
