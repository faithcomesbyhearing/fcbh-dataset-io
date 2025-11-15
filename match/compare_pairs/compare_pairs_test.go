package compare_pairs

import (
	"context"
	"gopkg.in/yaml.v3"
	"testing"
)

var request = `dataset_name: N2MRNWBT
pairs:
    - path: s3://dataset-io/GaryNGriswold/N2MRNWBT_v/00001/output/N2MRNWBT_v_audio_compare.json
      description: mms_adapter
    - path: s3://dataset-io/GaryNGriswold/N2MRNWBT_w/00002/output/N2MRNWBT_w_audio_compare.json
      description: wav2vec2_word
`

func TestComparePairs(t *testing.T) {
	ctx := context.Background()
	var p PairList
	err := yaml.Unmarshal([]byte(request), &p)
	if err != nil {
		t.Fatal(err)
	}
	status := ComparePairs(ctx, "", 0, p)
	if status != nil {
		t.Fatal(status)
	}
}

func TestN2MZJSIM_MAT12(t *testing.T) {
	var list PairList
	list.DatasetName = "N2MZJSIM"
	var asr2 OnePair
	asr2.Description = "N2MZJSIM_ASR2"
	asr2.Path = "s3://dataset-io/GaryNTest/N2MZJSIM/00001/output/N2MZJSIM_asr2.json"
	list.Pairs = append(list.Pairs, asr2)
	var mms OnePair
	mms.Description = "N2MZJSIM_MMS_only"
	mms.Path = "s3://dataset-io/GaryNTest/N2MZJSIM/00001/output/N2MZJSIM_audio_compare.json"
	list.Pairs = append(list.Pairs, mms)
	var adapter OnePair
	adapter.Description = "N2MZJSIM_adapter"
	adapter.Path = "s3://dataset-io/GaryNTest/N2MZJSIM/00003/output/N2MZJSIM_audio_compare.json"
	list.Pairs = append(list.Pairs, adapter)
	status := ComparePairs(context.TODO(), "MAT", 12, list)
	if status != nil {
		t.Fatal(status)
	}
	/// Add mms_adapter run
}
