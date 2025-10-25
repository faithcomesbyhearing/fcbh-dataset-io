package compare_pairs

import (
	"context"
	"encoding/json"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/match/diff"
	"github.com/xuri/excelize/v2"
	"os"
	"path/filepath"
	"strconv"
)

type PairList struct {
	DatasetName string    `yaml:"dataset_name"`
	Pairs       []OnePair `yaml:"pairs"`
}

type OnePair struct {
	Path        string `yaml:"path"` // required if pairMap nil
	Description string `yaml:"description"`
	PairMap     map[int]diff.Pair
}

func ComparePairs(ctx context.Context, pairs PairList) *log.Status {
	//var status *log.Status
	for i, p := range pairs.Pairs {
		if p.PairMap == nil {
			status := preparePairsMap(ctx, &p)
			if status != nil {
				return status
			}
			pairs.Pairs[i] = p
		}
	}
	rpt := NewExcelReport(ctx, pairs.DatasetName)
	for scr := 1; scr < 10000; scr++ {
		scrCount := 0
		for _, pair := range pairs.Pairs {
			one, ok := pair.PairMap[scr]
			if ok {
				status := rpt.generateLine(scrCount, pair.Description, one)
				if status != nil {
					return status
				}
				scrCount++
			}
		}
		if scrCount > 0 {

		}
	}
	status := rpt.writeFile()
	return status
}

func preparePairsMap(ctx context.Context, onePair *OnePair) *log.Status {
	var status *log.Status
	if onePair.Path == "" {
		return log.ErrorNoErr(ctx, 500, "Path to Pairs data is required.")
	}
	outputPath := filepath.Join(os.Getenv("FCBH_DATASET_TMP"), "pairs.json")
	status = input.DownloadFile(ctx, onePair.Path, outputPath)
	if status != nil {
		return status
	}
	bytes, err := os.ReadFile(outputPath)
	if err != nil {
		return log.ErrorNoErr(ctx, 500, "Failed to read file.")
	}
	pairs := make([]diff.Pair, 0)
	err = json.Unmarshal(bytes, &pairs)
	if err != nil {
		return log.ErrorNoErr(ctx, 500, "Failed to parse file.")
	}
	onePair.PairMap = make(map[int]diff.Pair)
	for i := range pairs {
		pairs[i].HTML = ""
		pairs[i].Diffs = nil
		onePair.PairMap[pairs[i].ScriptId()] = pairs[i]
	}
	return nil
}

type ExcelReport struct {
	ctx      context.Context
	file     *excelize.File
	filepath string
	lineNum  int
}

const SHEET1 = "Sheet1"

func NewExcelReport(ctx context.Context, title string) ExcelReport {
	var r ExcelReport
	r.ctx = ctx
	r.file = excelize.NewFile()
	r.filepath = title + ".xlsx"
	return r
}

func startGroup() {

}

func (r *ExcelReport) generateLine(scrCount int, description string, pair diff.Pair) *log.Status {
	var status *log.Status
	r.lineNum += 1
	if scrCount == 0 {
		status = r.writeCell("A", pair.Ref.Description())
		if status != nil {
			return status
		}
	}
	status = r.writeCell("B", description)
	if status != nil {
		return status
	}
	status = r.writeCell("C", "Text")
	if status != nil {
		return status
	}
	status = r.writeCell("D", pair.Base.Text)
	if status != nil {
		return status
	}
	r.lineNum += 1
	status = r.writeCell("C", "Audio")
	if status != nil {
		return status
	}
	status = r.writeCell("D", pair.Comp.Text)
	return status
}

func (r *ExcelReport) writeCell(col string, value string) *log.Status {
	cell := col + strconv.Itoa(r.lineNum)
	err := r.file.SetCellValue(SHEET1, cell, value)
	if err != nil {
		return log.Error(r.ctx, 500, err, "Unable to write cell.")
	}
	return nil
}

func (r *ExcelReport) writeLine(line []excelize.RichTextRun) *log.Status {
	r.lineNum += 1
	excelLine := "A" + strconv.Itoa(r.lineNum)
	err := r.file.SetCellRichText(SHEET1, excelLine, line)
	if err != nil {
		return log.Error(r.ctx, 500, err, "Failed to write excel line.")
	}
	return nil
}

func (r *ExcelReport) writeFile() *log.Status {
	err := r.file.SaveAs(r.filepath)
	if err != nil {
		return log.Error(r.ctx, 500, err, "Failed to save compare report")
	}
	//fmt.Println("Successfully created Compare.xlsx")
	return nil
}

/*


import "github.com/xuri/excelize/v2"

func main() {
	f := excelize.NewFile()

	// Example: A string where only certain parts are highlighted
	// Like marking "PASS" in green and "FAIL" in red within a longer text
	runs := []excelize.RichTextRun{
		{Text: "Test results: Module A ", Font: &excelize.Font{Color: "000000"}},
		{Text: "PASS", Font: &excelize.Font{Color: "008000", Bold: true}},  // Green + Bold
		{Text: ", Module B ", Font: &excelize.Font{Color: "000000"}},
		{Text: "FAIL", Font: &excelize.Font{Color: "FF0000", Bold: true}},  // Red + Bold
		{Text: ", Module C ", Font: &excelize.Font{Color: "000000"}},
		{Text: "PASS", Font: &excelize.Font{Color: "008000", Bold: true}},  // Green + Bold
		{Text: " - Review needed", Font: &excelize.Font{Color: "000000"}},
	}

	f.SetCellRichText("Sheet1", "A1", runs)

	// Another example - highlighting specific values in a log entry
	runs2 := []excelize.RichTextRun{
		{Text: "Connection from 192.168.1.100 status: ", Font: &excelize.Font{Color: "000000"}},
		{Text: "SUCCESS", Font: &excelize.Font{Color: "008000", Bold: true}},  // Green + Bold
		{Text: " at 10:34:22", Font: &excelize.Font{Color: "000000"}},
	}

	f.SetCellRichText("Sheet1", "A2", runs2)

	f.SaveAs("highlighted_text.xlsx")
}
*/
