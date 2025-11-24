package stdio_exec

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestStdioExec(t *testing.T) {
	ctx := context.Background()
	uromanPath := os.Getenv(`FCBH_UROMAN_EXE`)
	stdio1, status := NewStdioExec(ctx, uromanPath)
	//defer stdio1.Close()
	result, status2 := stdio1.Process("abc")
	fmt.Println("result:", result, status, status2)
	stdio1.Close()
}
