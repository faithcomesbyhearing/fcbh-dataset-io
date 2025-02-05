package diff

import (
	"context"
	"fmt"
	"testing"
)

func TestHTMLWriter(t *testing.T) {
	test := compareTest{baseDB: "N2ENGWEB", project: "N2ENGWEB_audio", expect: 2479}
	records, fileMap, status := runCompareTest(test)
	if status != nil {
		t.Fatal(status)
	}
	ctx := context.Background()
	writer, status := NewHTMLWriter(ctx, test.project)
	if status != nil {
		t.Fatal(status)
	}
	writer.WriteHeading(test.baseDB)
	for _, rec := range records {
		writer.WriteLine(rec)
	}
	filename := writer.WriteEnd(fileMap)
	fmt.Println("Filename:", filename)
}
