package decode_yaml

import (
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"reflect"
	"strings"
)

func (r *RequestDecoder) Validate(req *request.Request) {
	r.checkRequired(req)
	r.checkTestament(&req.Testament)
	r.checkAudioData(&req.AudioData, `AudioData`)
	r.checkTextData(&req.TextData, `TextData`)
	r.checkSpeechToText(&req.SpeechToText, `SpeechToText`)
	r.checkDetail(&req.Detail)
	r.checkTimestamps(&req.Timestamps, `Timestamps`)
	r.checkTraining(&req.Training, `Training`)
	r.checkAudioEncoding(&req.AudioEncoding, `AudioEncoding`)
	r.checkTextEncoding(&req.TextEncoding, `TextEncoding`)
	//checkCompare(req.Compare, &msgs)
	r.checkForOne(reflect.ValueOf(req.Compare.CompareSettings.DoubleQuotes), `DoubleQuotes`, true)
	r.checkForOne(reflect.ValueOf(req.Compare.CompareSettings.Apostrophe), `Apostrophe`, true)
	r.checkForOne(reflect.ValueOf(req.Compare.CompareSettings.Hyphen), `Hyphen`, true)
	r.checkForOne(reflect.ValueOf(req.Compare.CompareSettings.DiacriticalMarks), `DiscriticalMarks`, true)
}

func (r *RequestDecoder) checkRequired(req *request.Request) {
	if req.DatasetName == `` {
		r.errors = append(r.errors, `Required field dataset_name is empty`)
	}
	if req.BibleId == `` && req.LanguageISO == `` {
		r.errors = append(r.errors, `Required field bible_id: or language_iso: is empty`)
	}
	if req.Username == `` {
		r.errors = append(r.errors, `Required field username: is empty`)
	}
	req.DatasetName = strings.Replace(req.DatasetName, ` `, `_`, -1)
	if req.Compare.BaseDataset != `` {
		req.Compare.BaseDataset = strings.Replace(req.Compare.BaseDataset, ` `, `_`, -1)
	}
}

func (r *RequestDecoder) checkTestament(req *request.Testament) {
	if !req.OT && !req.NT && len(req.NTBooks) == 0 && len(req.OTBooks) == 0 {
		req.NT = true
	}
}

// checkAudioData Is checking that no more than one item is selected.
// if none are selected, it will set the default: NoAudio
func (r *RequestDecoder) checkAudioData(req *request.AudioData, fieldName string) {
	count := r.checkForOne(reflect.ValueOf(*req), fieldName, true)
	if count == 0 {
		req.NoAudio = true
	}
}

// checkTextData Is checking that no more than one item is selected.
// if none are selected, it will set the default: NoAudio
func (r *RequestDecoder) checkTextData(req *request.TextData, fieldName string) {
	count := r.checkForOne(reflect.ValueOf(*req), fieldName, true)
	if count == 0 {
		req.NoText = true
	}
}

func (r *RequestDecoder) checkSpeechToText(req *request.SpeechToText, fieldName string) {
	//whisper := req.Whisper
	count := r.checkForOne(reflect.ValueOf(*req), fieldName, true)
	if count == 0 {
		req.NoSpeechToText = true
	}
}

func (r *RequestDecoder) checkDetail(req *request.Detail) {
	if !req.Lines && !req.Words {
		req.Lines = true
	}
}

func (r *RequestDecoder) checkTimestamps(req *request.Timestamps, fieldName string) {
	count := r.checkForOne(reflect.ValueOf(*req), fieldName, true)
	if count == 0 {
		req.NoTimestamps = true
	}
}

func (r *RequestDecoder) checkTraining(req *request.Training, fieldName string) {
	count := r.checkForOne(reflect.ValueOf(*req), fieldName, false)
	if count == 0 {
		req.NoTraining = true
	}
}

func (r *RequestDecoder) checkAudioEncoding(req *request.AudioEncoding, fieldName string) {
	count := r.checkForOne(reflect.ValueOf(*req), fieldName, true)
	if count == 0 {
		req.NoEncoding = true
	}
}

func (r *RequestDecoder) checkTextEncoding(req *request.TextEncoding, fieldName string) {
	count := r.checkForOne(reflect.ValueOf(*req), fieldName, true)
	if count == 0 {
		req.NoEncoding = true
	}
}

func (r *RequestDecoder) checkForOne(structVal reflect.Value, fieldName string, recurse bool) int {
	var errorCount int
	var wasSet []string
	r.checkForOneRecursive(structVal, &wasSet, recurse)
	errorCount += len(wasSet)
	if len(wasSet) > 1 {
		msg := `Only 1 field can be set on ` + fieldName + `: ` + strings.Join(wasSet, `,`)
		r.errors = append(r.errors, msg)
	}
	return errorCount
}

func (r *RequestDecoder) checkForOneRecursive(sVal reflect.Value, wasSet *[]string, recurse bool) {
	for i := 0; i < sVal.NumField(); i++ {
		field := sVal.Field(i)
		if field.Kind() == reflect.String {
			if field.String() != `` {
				*wasSet = append(*wasSet, sVal.Type().Field(i).Name)
			}
		} else if field.Kind() == reflect.Bool {
			if field.Bool() {
				*wasSet = append(*wasSet, sVal.Type().Field(i).Name)
			}
		} else if field.Kind() == reflect.Int {
			if field.Int() != 0 && len(*wasSet) == 0 {
				*wasSet = append(*wasSet, sVal.Type().Field(i).Name)
			}
		} else if field.Kind() == reflect.Struct {
			if recurse {
				r.checkForOneRecursive(field, wasSet, recurse)
			}
		} else {
			msg := sVal.Type().Field(i).Name + ` has unexpected type ` + field.Type().Name()
			r.errors = append(r.errors, msg)
		}
	}
}
