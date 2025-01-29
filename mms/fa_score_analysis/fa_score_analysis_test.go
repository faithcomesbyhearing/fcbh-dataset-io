package fa_score_analysis

import (
	"context"
	"dataset/db"
	"dataset/decode_yaml/request"
	"fmt"
	"testing"
)

func TestFAScoreAnalysis(t *testing.T) {
	ctx := context.Background()
	user := request.GetTestUser()
	list := []string{"N2KTB_ESB", "N2CFM_BSM", "N2CHF_TBL", "N2CUL_MNT"}
	for _, database := range list {
		conn, status := db.NewerDBAdapter(ctx, false, user, database)
		if status != nil {
			t.Fatal(status)
		}
		output, status := FAScoreAnalysis(conn)
		if status != nil {
			t.Fatal(status)
		}
		conn.Close()
		fmt.Println(output)
	}
}
