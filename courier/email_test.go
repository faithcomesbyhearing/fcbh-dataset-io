package courier

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestGoMailSendMail(t *testing.T) {
	ctx := context.Background()
	recipients := []string{"gary@shortsands.com"}
	subject := "TestSESSendMail"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	msg := "TestSESSendMail " + strconv.Itoa(r.Intn(100))
	fmt.Println("SENT:", msg)
	err := GoMailSendMail(ctx, recipients, subject, msg, []string{})
	if err != nil {
		t.Fatal(err)
	}
}
