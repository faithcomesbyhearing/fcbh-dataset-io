package tests

import "testing"

const vesselAsrTest = `is_new: yes
dataset_name: P2SDOCAK
language_iso: sdo
username: vessel
notify_ok: [gary@shortsands.com]
notify_err: [gary@shortsands.com]
testament:
  nt: yes
text_data:
  aws_s3: s3://dataset-vessel/vessel/P2SDOCAK_10221709/P2SDOCAK Text/XLSX/Bidayuh Serian (LUK)_P2SDOCAK_LineBased.xlsx
audio_data:
  aws_s3: s3://dataset-vessel/vessel/P2SDOCAK_10221709/P2SDOCAK Line VOX/*.wav
speech_to_text:
  mms_asr: yes
`

func TestAsrVessel(t *testing.T) {
	var yaml = vesselAsrTest
	DirectSqlTest(yaml, []SqliteTest{}, t)
}
