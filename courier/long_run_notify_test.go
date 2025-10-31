package courier

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"testing"
	"time"
)

func TestLongRunNotify(t *testing.T) {
	ctx := context.Background()
	log.SetOutput("stdout")
	const yamlRequest = `is_new: Y
username: Sam_I_Am
dataset_name: Test_Dataset
language_iso: eng
notify_ok: [gary@shortsands.com, sqs/vessel]
notify_err: [gary@shortsands.com, sqs/vessel]
`
	reqDecoder := decode_yaml.NewRequestDecoder(ctx)
	request, status := reqDecoder.Process([]byte(yamlRequest))
	if status != nil {
		t.Fatal(status)
	}
	LongRunNotify(ctx, request)
	time.Sleep(1 * time.Minute)
}
