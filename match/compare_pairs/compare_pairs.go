package compare_pairs

import (
	"context"
	"encoding/json"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/input"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/match/diff"
	"github.com/sergi/go-diff/diffmatchpatch"
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
	err := rpt.setStyle()
	if err != nil {
		return log.Error(ctx, 500, err, "Could not set style")
	}
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
		onePair.PairMap[pairs[i].ScriptId()] = pairs[i]
	}
	return nil
}

type ExcelReport struct {
	ctx         context.Context
	file        *excelize.File
	filepath    string
	styleId     int
	colDStyleId int
	lineNum     int
}

const SHEET1 = "Sheet1"

func NewExcelReport(ctx context.Context, title string) ExcelReport {
	var r ExcelReport
	r.ctx = ctx
	r.file = excelize.NewFile()
	r.filepath = title + ".xlsx"
	return r
}

func (r *ExcelReport) setStyle() *log.Status {
	var err error
	r.styleId, err = r.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   12,
			Family: "Calibri",
			Color:  "#000000",
		},
	})
	if err != nil {
		return log.Error(r.ctx, 500, err, "Failed to create new style.")
	}
	_ = r.file.SetColWidth(SHEET1, "A", "A", 9)
	_ = r.file.SetColWidth(SHEET1, "B", "B", 14)
	_ = r.file.SetColWidth(SHEET1, "C", "C", 6)
	_ = r.file.SetColWidth(SHEET1, "D", "D", 120) // Adjust as needed

	r.colDStyleId, err = r.file.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			WrapText: true,
			Vertical: "top", // Align to top of cell
		},
		Font: &excelize.Font{
			Size:   12,
			Family: "Calibri",
			Color:  "#000000",
		},
	})
	if err != nil {
		return log.Error(r.ctx, 500, err, "Failed to create new style.")
	}
	return nil
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
	if status != nil {
		return status
	}
	r.lineNum += 1
	status = r.writeCell("C", "Diff")
	if status != nil {
		return status
	}
	status = r.writeLine("D", r.generateDiffLine(pair.Diffs))
	return status
}

func (r *ExcelReport) generateDiffLine(diffs []diffmatchpatch.Diff) []excelize.RichTextRun {
	var result []excelize.RichTextRun
	for _, diff := range diffs {
		var item excelize.RichTextRun
		switch diff.Type {
		case diffmatchpatch.DiffEqual: // In both text and audio
			item = excelize.RichTextRun{Text: diff.Text, Font: &excelize.Font{
				Size:   12,
				Family: "Calibri",
				Color:  "#000000"}}
		case diffmatchpatch.DiffDelete: // In text only red
			item = excelize.RichTextRun{Text: diff.Text, Font: &excelize.Font{
				Size:   12,
				Family: "Calibri",
				Color:  "#FF0000"}}
		case diffmatchpatch.DiffInsert: // In audio only green
			item = excelize.RichTextRun{Text: diff.Text, Font: &excelize.Font{
				Size:   12,
				Family: "Calibri",
				Color:  "#008000"}}
		}
		result = append(result, item)
	}
	return result
}

func (r *ExcelReport) writeCell(col string, value string) *log.Status {
	cell := col + strconv.Itoa(r.lineNum)
	err := r.file.SetCellValue(SHEET1, cell, value)
	if err != nil {
		return log.Error(r.ctx, 500, err, "Unable to write cell.")
	}
	return nil
}

func (r *ExcelReport) writeLine(col string, line []excelize.RichTextRun) *log.Status {
	cell := col + strconv.Itoa(r.lineNum)
	err := r.file.SetCellRichText(SHEET1, cell, line)
	if err != nil {
		return log.Error(r.ctx, 500, err, "Failed to write excel line.")
	}
	return nil
}

func (r *ExcelReport) writeFile() *log.Status {
	lastCell := "C" + strconv.Itoa(r.lineNum)
	err := r.file.SetCellStyle(SHEET1, "A1", lastCell, r.styleId)
	if err != nil {
		return log.Error(r.ctx, 500, err, "Failed to set styles for A-C.")
	}
	lastCell = "D" + strconv.Itoa(r.lineNum)
	err = r.file.SetCellStyle(SHEET1, "D1", lastCell, r.colDStyleId)
	if err != nil {
		return log.Error(r.ctx, 500, err, "Failed to set styles for D.")
	}
	err = r.file.SaveAs(r.filepath)
	if err != nil {
		return log.Error(r.ctx, 500, err, "Failed to save compare report")
	}
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
