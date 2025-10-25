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
	status := ComparePairs(ctx, p)
	if status != nil {
		t.Fatal(status)
	}
}
