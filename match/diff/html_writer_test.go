package diff

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestHTMLWriter(t *testing.T) {
	test := compareTest{baseDB: "N2ENGWEB", project: "N2ENGWEB_audio", expect: 2479}
	records, fileMap, languageISO, status := runCompareTest(test)
	if status != nil {
		t.Fatal(status)
	}
	ctx := context.Background()
	writer := NewHTMLWriter(ctx, test.project)
	filename, status := writer.WriteReport(test.baseDB, records, languageISO, fileMap)
	if status != nil {
		t.Fatal(status)
	}
	newPath := filepath.Join("../", filepath.Base(filename))
	err := os.Rename(filename, newPath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Filename:", newPath)
}
