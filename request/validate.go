package request

import (
	"reflect"
	"strings"
)

func (r *RequestDecoder) Validate(req *Request) {
	r.checkRequired(req)
	r.checkTestament(&req.Testament)
	r.checkAudioData(&req.AudioData, `AudioData`)
	r.checkTextData(&req.TextData, `TextData`)
	r.checkSpeechToText(&req.SpeechToText, `SpeechToText`)
	r.checkDetail(&req.Detail)
	r.checkTimestamps(&req.Timestamps, `Timestamps`)
	r.checkAudioEncoding(&req.AudioEncoding, `AudioEncoding`)
	r.checkTextEncoding(&req.TextEncoding, `TextEncoding`)
	//checkCompare(req.Compare, &msgs)
	r.checkForOne(reflect.ValueOf(req.Compare.CompareSettings.DoubleQuotes), `DoubleQuotes`)
	r.checkForOne(reflect.ValueOf(req.Compare.CompareSettings.Apostrophe), `Apostrophe`)
	r.checkForOne(reflect.ValueOf(req.Compare.CompareSettings.Hyphen), `Hyphen`)
	r.checkForOne(reflect.ValueOf(req.Compare.CompareSettings.DiacriticalMarks), `DiscriticalMarks`)
}

func (r *RequestDecoder) checkRequired(req *Request) {
	if req.DatasetName == `` {
		r.errors = append(r.errors, `Required field dataset_name is empty`)
	}
	if req.BibleId == `` {
		r.errors = append(r.errors, `Required field bible_id: is empty`)
	}
	if req.Username == `` {
		r.errors = append(r.errors, `Required field username: is empty`)
	}
	if req.Email == `` {
		r.errors = append(r.errors, `Required field email: is empty`)
	}
	req.DatasetName = strings.Replace(req.DatasetName, ` `, `_`, -1)
	if req.Compare.BaseDataset != `` {
		req.Compare.BaseDataset = strings.Replace(req.Compare.BaseDataset, ` `, `_`, -1)
	}
}

func (r *RequestDecoder) checkTestament(req *Testament) {
	if !req.OT && !req.NT && len(req.NTBooks) == 0 && len(req.OTBooks) == 0 {
		req.NT = true
	}
}

// checkAudioData Is checking that no more than one item is selected.
// if none are selected, it will set the default: NoAudio
func (r *RequestDecoder) checkAudioData(req *AudioData, fieldName string) {
	count := r.checkForOne(reflect.ValueOf(*req), fieldName)
	if count == 0 {
		req.NoAudio = true
	}
}

// checkTextData Is checking that no more than one item is selected.
// if none are selected, it will set the default: NoAudio
func (r *RequestDecoder) checkTextData(req *TextData, fieldName string) {
	count := r.checkForOne(reflect.ValueOf(*req), fieldName)
	if count == 0 {
		req.NoText = true
	}
}

func (r *RequestDecoder) checkSpeechToText(req *SpeechToText, fieldName string) {
	//whisper := req.Whisper
	count := r.checkForOne(reflect.ValueOf(*req), fieldName)
	if count == 0 {
		req.NoSpeechToText = true
	}
}

func (r *RequestDecoder) checkDetail(req *Detail) {
	if !req.Lines && !req.Words {
		req.Lines = true
	}
}

func (r *RequestDecoder) checkTimestamps(req *Timestamps, fieldName string) {
	count := r.checkForOne(reflect.ValueOf(*req), fieldName)
	if count == 0 {
		req.NoTimestamps = true
	}
}

func (r *RequestDecoder) checkAudioEncoding(req *AudioEncoding, fieldName string) {
	count := r.checkForOne(reflect.ValueOf(*req), fieldName)
	if count == 0 {
		req.NoEncoding = true
	}
}

func (r *RequestDecoder) checkTextEncoding(req *TextEncoding, fieldName string) {
	count := r.checkForOne(reflect.ValueOf(*req), fieldName)
	if count == 0 {
		req.NoEncoding = true
	}
}

func (r *RequestDecoder) checkForOne(structVal reflect.Value, fieldName string) int {
	var errorCount int
	var wasSet []string
	r.checkForOneRecursive(structVal, &wasSet)
	errorCount += len(wasSet)
	if len(wasSet) > 1 {
		msg := `Only 1 field can be set on ` + fieldName + `: ` + strings.Join(wasSet, `,`)
		r.errors = append(r.errors, msg)
	}
	return errorCount
}

func (r *RequestDecoder) checkForOneRecursive(sVal reflect.Value, wasSet *[]string) {
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
		} else if field.Kind() == reflect.Struct {
			r.checkForOneRecursive(field, wasSet)
		} else {
			msg := sVal.Type().Field(i).Name + ` has unexpected type ` + field.Type().Name()
			r.errors = append(r.errors, msg)
		}
	}
}
