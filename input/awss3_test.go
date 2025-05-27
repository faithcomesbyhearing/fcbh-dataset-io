package input

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"os"
	"path/filepath"
	"testing"
)

func TestAWSS3USX1(t *testing.T) {
	AWSS3USX(`s3://dbp-prod/text/ENGWEB/ENGWEBN_ET-usx/*.usx`, t)
}
func TestAWSS3USX2(t *testing.T) {
	AWSS3USX(`s3://dbp-prod/text/ENGWEB/ENGWEBN_ET-usx`, t)
}
func TestAWSS3USX3(t *testing.T) {
	AWSS3USX(`s3://dbp-prod/text/ENGWEB/ENGWEBN_ET-usx/`, t)
}

func AWSS3USX(key string, t *testing.T) {
	ctx := context.Background()
	//key := `s3://dbp-prod/text/ENGWEB/ENGWEBN_ET-usx/*.usx`
	testament := request.Testament{NT: true}
	testament.BuildBookMaps()
	directory := filepath.Join(os.Getenv(`FCBH_DATASET_FILES`), `ENGWEB`, `ENGWEBN_ET-usx`)
	_ = os.Remove(filepath.Join(directory, `040MAT.usx`))
	_ = os.Remove(filepath.Join(directory, `041MRK.usx`))
	_ = os.Remove(filepath.Join(directory, `042LUK.usx`))
	_ = os.Remove(filepath.Join(directory, `043JHN.usx`))
	files, status := AWSS3Input(ctx, key)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 27 {
		t.Error(`len(files) should be 27`, len(files))
	}
	if files[2].MediaId != `ENGWEBN_ET-usx` {
		t.Error(`files[0].MediaId should be ENGWEBN_ET-usx`, files[2].MediaId)
	}
	if files[2].MediaType != `text_usx` {
		t.Error(`files[0].MediaType should be text_usx`, files[2].MediaType)
	}
}

func TestAWSAudio1(t *testing.T) {
	AWSS3Audio(`s3://dbp-prod/audio/ENGWEB/ENGWEBN2DA/*.mp3`, 260, t)
}

func TestAWSAudio2(t *testing.T) {
	AWSS3Audio(`s3://dbp-prod/audio/ENGWEB/ENGWEBN2DA/`, 267, t)
}

func TestAWSAudio3(t *testing.T) {
	AWSS3Audio(`s3://dbp-prod/audio/ENGWEB/ENGWEBN2DA/`, 267, t)
}

func AWSS3Audio(key string, expect int, t *testing.T) {
	ctx := context.Background()
	testament := request.Testament{NT: true}
	testament.BuildBookMaps()
	directory := filepath.Join(os.Getenv(`FCBH_DATASET_FILES`), `ENGWEB`, `ENGWEBN2DA`)
	_ = os.Remove(filepath.Join(directory, `B27___17_Revelation__ENGWEBN2DA.mp3`))
	_ = os.Remove(filepath.Join(directory, `B27___18_Revelation__ENGWEBN2DA.mp3`))
	_ = os.Remove(filepath.Join(directory, `B27___19_Revelation__ENGWEBN2DA.mp3`))
	_ = os.Remove(filepath.Join(directory, `B27___20_Revelation__ENGWEBN2DA.mp3`))
	files, status := AWSS3Input(ctx, key)
	if status != nil {
		t.Error(status)
	}
	if len(files) != expect {
		t.Error(`len(files) should be `, expect, len(files))
	}
	if files[2].MediaId != `ENGWEBN2DA` {
		t.Error(`files[0].MediaId should be ENGWEBN2DA`, files[2].MediaId)
	}
	if files[2].MediaType != `audio` {
		t.Error(`files[0].MediaType should be audio`, files[2].MediaType)
	}
}

func NotYetTestAWSS3Zip(t *testing.T) {
	ctx := context.Background()
	key := `s3://dbp-prod/audio/ENGWEB/ENGWEBN2DA/ENGWEBN2DA.zip`
	testament := request.Testament{NT: true}
	testament.BuildBookMaps()
	//directory := filepath.Join(os.Getenv(`FCBH_DATASET_FILES`), `ENGWEB`, `ENGWEBN2DA`)
	//os.Remove(filepath.Join(directory, `040MAT.usx`))
	//os.Remove(filepath.Join(directory, `041MRK.usx`))
	//os.Remove(filepath.Join(directory, `042LUK.usx`))
	//os.Remove(filepath.Join(directory, `043JHN.usx`))
	files, status := AWSS3Input(ctx, key)
	if status != nil {
		t.Error(status)
	}
	if len(files) != 1 {
		t.Error(`len(files) should be 260`, len(files))
	}
}
