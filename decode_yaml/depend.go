package decode_yaml

import "github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"

func (r *RequestDecoder) Prereq(req *request.Request) {
	if req.Timestamps.MMSAlign {
		req.Detail.Words = true
	}
}

func (r *RequestDecoder) Depend(req request.Request) {
	if req.Database.AWSS3 != "" {
		if req.IsNew {
			r.errors = append(r.errors, `When database.aws_s3 is set, is_new must be false`)
		}
	}
	if !req.Timestamps.NoTimestamps {
		if req.AudioData.NoAudio {
			r.errors = append(r.errors, `Timestamps are requested, but there is no audio`)
		}
		if req.TextData.NoText {
			r.errors = append(r.errors, `Timestamps are requested, but there is no text`)
			// The need for text is not a real requirement, but the system is coded to store timestamps
			// in the scripts table, and it cannot do this unless there is text.  If this becomes
			// a problem the system could be changed to insert timestamp data without text.
		}
	}
	if req.Timestamps.Aeneas || req.Timestamps.MMSFAVerse || req.Timestamps.MMSAlign {
		if req.TextData.NoText {
			r.errors = append(r.errors, `Timestamp estimation requested, but there is no text data`)
		}
	}
	if !req.TextEncoding.NoEncoding {
		if req.TextData.NoText {
			r.errors = append(r.errors, `Text encoding requested, but there is no text data`)
		}
	}
	if !req.SpeechToText.NoSpeechToText {
		if req.AudioData.NoAudio {
			r.errors = append(r.errors, `Speech to Text is requested, but there is no audio`)
		}
		if req.Timestamps.NoTimestamps {
			r.errors = append(r.errors, `Speech to Text is requested, but there are no timestamps`)
		}
	}
	if req.AudioEncoding.MFCC {
		if req.Timestamps.NoTimestamps {
			r.errors = append(r.errors, `MFCC's are requested', but there are no timestamps`)
		}
	}
	if req.AudioProof.HTMLReport {
		if req.IsNew {
			if !req.Timestamps.MMSAlign {
				r.errors = append(r.errors, `AudioProof is requested, but there is no mms_align`)
			}
			if !req.SpeechToText.MMS {
				r.errors = append(r.errors, `AudioProof is requested, but there is no MMS_ASR`)
			}
		} else {
			if req.AudioProof.BaseDataset == "" {
				r.errors = append(r.errors, `AudioProof is requested on existing dataset, but there is no BaseDataset`)
			}
		}
	}
}
