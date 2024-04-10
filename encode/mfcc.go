package encode

import (
	"bytes"
	"context"
	"dataset"
	"dataset/db"
	log "dataset/logger"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

type MFCC struct {
	ctx       context.Context
	conn      db.DBAdapter
	bibleId   string
	audioFSId string
}

func NewMFCC(ctx context.Context, conn db.DBAdapter, bibleId string, audioFSId string) MFCC {
	var m MFCC
	m.ctx = ctx
	m.conn = conn
	m.bibleId = bibleId
	m.audioFSId = audioFSId
	return m
}

func (m *MFCC) Process(detail dataset.TextDetailType, numMFCC int) dataset.Status {
	var status dataset.Status
	audioFiles, status := ReadDirectory(m.ctx, m.bibleId, m.audioFSId)
	if status.IsErr {
		return status
	}
	for _, audioFile := range audioFiles {
		var mfccResp MFCCResp
		mfccResp, status = m.executeLibrosa(audioFile, numMFCC)
		if status.IsErr {
			return status
		}
		bookId, chapterNum, status := ParseFilename(m.ctx, audioFile)
		if status.IsErr {
			return status
		}
		if detail == dataset.LINES || detail == dataset.BOTH {
			status = m.processScripts(mfccResp, bookId, chapterNum)
			if status.IsErr {
				return status
			}
		} else if detail == dataset.WORDS || detail == dataset.BOTH {
			status = m.processWords(mfccResp, bookId, chapterNum)
			if status.IsErr {
				return status
			}
		}
	}
	return status
}

type MFCCResp struct {
	AudioFile  string      `json:"input_file"`
	SampleRate float64     `json:"sample_rate"`
	HopLength  float64     `json:"hop_length"`
	FrameRate  float64     `json:"frame_rate"`
	Shape      []int       `json:"mfcc_shape"`
	Type       string      `json:"mfcc_type"`
	MFCC       [][]float32 `json:"mfccs"`
}

func (m *MFCC) executeLibrosa(audioFile string, numMFCC int) (MFCCResp, dataset.Status) {
	var result MFCCResp
	var status dataset.Status
	pythonPath := "python3"
	cmd := exec.Command(pythonPath, `mfcc_librosa.py`, audioFile, strconv.Itoa(numMFCC))
	fmt.Println(cmd.String())
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	if err != nil {
		status = log.Error(m.ctx, 500, err, `Error executing mfcc_librosa.py`)
		return result, status
	}
	if stderrBuf.Len() > 0 {
		status = log.ErrorNoErr(m.ctx, 500, `mfcc_librosa.py stderr:`, stderrBuf.String())
		return result, status
	}
	if stdoutBuf.Len() == 0 {
		status = log.ErrorNoErr(m.ctx, 500, `mfcc_librosa.py has no output.`)
		return result, status
	}
	err = json.Unmarshal(stdoutBuf.Bytes(), &result)
	if err != nil {
		status = log.Error(m.ctx, 500, err, `Error parsing json from librosa`)
	}
	return result, status
}

func (m *MFCC) processScripts(mfcc MFCCResp, bookId string, chapterNum int) dataset.Status {
	var status dataset.Status
	timestamps, status := m.conn.SelectScriptTimestamps(bookId, chapterNum)
	mfccs := m.segmentMFCC(timestamps, mfcc)
	status = m.conn.InsertScriptMFCCS(mfccs)
	return status
}

func (m *MFCC) processWords(mfcc MFCCResp, bookId string, chapterNum int) dataset.Status {
	var status dataset.Status
	timestamps, status := m.conn.SelectWordTimestamps(bookId, chapterNum)
	mfccs := m.segmentMFCC(timestamps, mfcc)
	status = m.conn.InsertWordMFCCS(mfccs)
	return status
}

func (m *MFCC) segmentMFCC(timestamps []db.Timestamp, mfcc MFCCResp) []db.MFCC {
	var mfccs []db.MFCC
	for _, ts := range timestamps {
		startIndex := int(ts.BeginTS*mfcc.FrameRate + 0.5)
		endIndex := int(ts.EndTS*mfcc.FrameRate + 0.5)
		segment := mfcc.MFCC[startIndex:endIndex][:]
		var mfcc db.MFCC
		mfcc.Id = ts.Id
		mfcc.Rows = len(segment)
		if mfcc.Rows == 0 {
			mfcc.Cols = 0
		} else {
			mfcc.Cols = len(segment[0])
		}
		mfcc.MFCC = segment
		mfccs = append(mfccs, mfcc)
	}
	return mfccs
}