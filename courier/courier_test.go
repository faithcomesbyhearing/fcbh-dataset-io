package courier

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"os"
	"testing"
	"time"
)

const runBucketTest = `is_new: yes
dataset_name: MyProject
bible_id: ENGWEB
username: GaryNTest
email: gary@shortsands.com
output_file: abc/my_project.csv
`

func TestCourier(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	b := NewCourier(ctx, []byte(runBucketTest))
	b.IsUnitTest = true
	if b.username != "GaryNTest" {
		t.Error("Username should be GaryNTest, it is: ", b.username)
	}
	if len(b.username) != 9 {
		t.Error("Username should be 9 characters")
	}
	if b.dataset != "MyProject" {
		t.Error("Project should be MyProject, it is:", b.dataset)
	}
	b.AddLogFile(os.Getenv("FCBH_DATASET_LOG_FILE"))
	database1, status := db.NewerDBAdapter(ctx, true, b.username, "TestCourier1")
	if status != nil {
		t.Fatal(status)
	}
	b.AddDatabase(database1)
	database2, status := db.NewerDBAdapter(ctx, true, b.username, "TestCourier2")
	if status != nil {
		t.Fatal(status)
	}
	b.AddDatabase(database2)
	b.AddOutput("../tests/02__plain_text_edit_script.csv")
	b.AddOutput("../tests/02__plain_text_edit_script.json")
	status = b.PersistToBucket()
	if status != nil {
		t.Fatal(status)
	}
	duration := time.Since(start)
	status = b.Notification(status, duration)
	status = log.ErrorNoErr(ctx, 400, "Test Error")
	status = b.Notification(status, duration)
}
