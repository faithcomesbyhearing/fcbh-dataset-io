package gen_yaml

import (
	"fmt"
	"testing"
)

func TestFindFilesets(t *testing.T) {
	bibles, status := FindFilesets()
	if status != nil {
		t.Fatal(status)
	}
	fmt.Println("num bibles, ", len(bibles))
}

func TestGenerateYaml(t *testing.T) {
	bibles, status := FindFilesets()
	if status != nil {
		t.Fatal(status)
	}
	status = GenerateYaml(bibles)
	if status != nil {
		t.Fatal(status)
	}

	//fmt.Println("num filesets, ", len(filesets))

}

func TestSizeRecognition(t *testing.T) {
	bibles, status := ReadBibles()
	if status != nil {
		t.Fatal(status)
	}
	for _, bible := range bibles {
		for _, fs := range bible.DbpProd.Filesets {
			size := ReduceSize(fs.Size)
			var testament string
			if len(fs.Id) > 6 {
				testament = fs.Id[6:7]
			} else {
				testament = "?"
			}
			if size != testament {
				fmt.Println(fs.Id, fs.Size, size, testament)
			}
		}
	}
}
