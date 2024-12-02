package match

import (
	"context"
	"dataset"
	log "dataset/logger"
	"os"
	"strconv"
	"strings"
	"time"
)

type AlignWriter struct {
	ctx         context.Context
	datasetName string
	out         *os.File
	lineNum     int
	critErrors  int
	questErrors int
	critGaps    int
	questGaps   int
}

func NewAlignWriter(ctx context.Context) AlignWriter {
	var a AlignWriter
	a.ctx = ctx
	return a
}

func (a *AlignWriter) WriteReport(datasetName string, verses []FAverse) (string, dataset.Status) {
	var filename string
	var status dataset.Status
	var err error
	a.datasetName = datasetName
	a.out, err = os.CreateTemp(os.Getenv(`FCBH_DATASET_TMP`), datasetName+"_*.html")
	if err != nil {
		return filename, log.Error(a.ctx, 500, err, `Error creating output file for align writer`)
	}
	a.WriteHeading()
	for _, vers := range verses {
		a.WriteVerse(vers)
	}
	a.WriteEnd()
	return a.out.Name(), status
}

func (a *AlignWriter) WriteHeading() {
	head := `<!DOCTYPE html>
<html>
 <head>
  <meta charset="utf-8">
  <title>Alignment Error Report</title>
`
	_, _ = a.out.WriteString(head)
	_, _ = a.out.WriteString(`<link rel="stylesheet" type="text/css" href="https://cdn.datatables.net/1.10.21/css/jquery.dataTables.css">`)
	_, _ = a.out.WriteString("</head><body>\n")
	_, _ = a.out.WriteString(`<h2 style="text-align:center">Audio to Text Alignment Report For `)
	_, _ = a.out.WriteString(a.datasetName)
	_, _ = a.out.WriteString("</h2>\n")
	_, _ = a.out.WriteString(`<h3 style="text-align:center">`)
	_, _ = a.out.WriteString(time.Now().Format(`Mon Jan 2 2006 03:04:05 pm MST`))
	_, _ = a.out.WriteString("</h3>\n")
	checkbox := `<div style="text-align: center; margin: 10px;">
		<input type="checkbox" id="hideVerse0" checked><label for="hideVerse0">Hide Headings</label>
	</div>
`
	_, _ = a.out.WriteString(checkbox)
	table := `<table id="diffTable" class="display">
    <thead>
    <tr>
        <th>Line</th>
		<th>Combined<br>Error</th>
        <th>Align<br>Error</th>
		<th>End<br>Gap</th>
		<th>Start<br>TS</th>
		<th>End<br>TS</th>
        <th>Ref</th>
		<th>Script</th>
    </tr>
    </thead>
    <tbody>
`
	_, _ = a.out.WriteString(table)
}

func (a *AlignWriter) WriteVerse(verse FAverse) {
	a.lineNum++
	_, _ = a.out.WriteString("<tr>\n")
	a.writeCell(strconv.Itoa(a.lineNum))
	score := verse.critScore + verse.questScore + 5.0*verse.endTSDiff
	a.writeCell(strconv.FormatFloat(score, 'f', 1, 64))
	a.writeCell(strconv.FormatFloat(verse.questScore, 'f', 1, 64))
	a.writeCell(strconv.FormatFloat(verse.endTSDiff, 'f', 2, 64))
	a.writeCell(a.minSecFormat(verse.beginTS))
	a.writeCell(a.minSecFormat(verse.endTS))
	a.writeCell(verse.bookId + ` ` + strconv.Itoa(verse.chapter) + `:` + verse.verseStr)
	var critical = a.createHighlightList(verse.critWords, len(verse.words))
	var question = a.createHighlightList(verse.questWords, len(verse.words))
	var text []string
	for i, wd := range verse.words {
		if critical[i] {
			a.critErrors++
			text = append(text, `<span class="red-box">`+wd.Text+`</span>`)
		} else if question[i] {
			a.questErrors++
			text = append(text, `<span class="yellow-box">`+wd.Text+`</span>`)
		} else {
			text = append(text, wd.Text)
		}
	}
	var diff int
	if verse.critDiff || verse.questDiff {
		diff = int((verse.endTSDiff - 0.9) * 10.0) // subract 1sec, and convert to 1/10 sec per char
		if diff > 0 {
			spaces := strings.Repeat("&nbsp;", diff)
			if verse.critDiff {
				a.critGaps++
				text = append(text, `<span class="red-box">`+spaces+`</span>`)
			} else if verse.questDiff {
				a.questGaps++
				text = append(text, `<span class="yellow-box">`+spaces+`</span>`)
			}
		}
	}
	a.writeCell(strings.Join(text, " "))
	_, _ = a.out.WriteString("</tr>\n")
}

func (a *AlignWriter) createHighlightList(indexes []int, length int) []bool {
	var list = make([]bool, length)
	for _, index := range indexes {
		list[index] = true
	}
	return list
}

func (a *AlignWriter) writeCell(content string) {
	_, _ = a.out.WriteString(`<td>`)
	_, _ = a.out.WriteString(content)
	_, _ = a.out.WriteString(`</td>`)
}

func (a *AlignWriter) WriteEnd() {
	table := `</tbody>
	</table>
`
	_, _ = a.out.WriteString(table)
	_, _ = a.out.WriteString(`<p>Lines with critical errors `)
	_, _ = a.out.WriteString(strconv.Itoa(a.critErrors))
	_, _ = a.out.WriteString(`</p>`)
	_, _ = a.out.WriteString(`<p>Lines with questionable errors `)
	_, _ = a.out.WriteString(strconv.Itoa(a.questErrors))
	_, _ = a.out.WriteString("</p>\n")
	_, _ = a.out.WriteString(`<p>Lines with large end-of-verse gaps `)
	_, _ = a.out.WriteString(strconv.Itoa(a.critGaps))
	_, _ = a.out.WriteString(`</p>`)
	_, _ = a.out.WriteString(`<p>Lines with smaller end-of-verse gaps `)
	_, _ = a.out.WriteString(strconv.Itoa(a.questGaps))
	_, _ = a.out.WriteString("</p>\n")
	_, _ = a.out.WriteString(`<script type="text/javascript" src="https://code.jquery.com/jquery-3.5.1.js"></script>`)
	_, _ = a.out.WriteString("\n")
	_, _ = a.out.WriteString(`<script type="text/javascript" src="https://cdn.datatables.net/1.10.21/js/jquery.dataTables.js"></script>`)
	_, _ = a.out.WriteString("\n")
	style := `<style>
	.dataTables_length select {
	width: auto;
	display: inline-block;
	padding: 5px;
		margin-left: 5px;
		border-radius: 4px;
	border: 1px solid #ccc;
	}
	.dataTables_filter input {
		width: auto;
		display: inline-block;
		padding: 5px;
		border-radius: 4px;
		border: 1px solid #ccc;
	}
	.dataTables_wrapper .dataTables_length, .dataTables_wrapper .dataTables_filter {
		margin-bottom: 20px;
	}
	.red-box { 
		background-color: rgba(255, 0, 0, 0.4);
		padding: 0 3px; /* 0 top/bottom, 3px left/right */ 
		border-radius: 3px; /* rounded corners */ 
	} 
	.yellow-box { 
		background-color: rgba(255, 255, 0, 0.8);
		padding: 0 3px; 
		border-radius: 3px; 
	}
	</style>
`
	_, _ = a.out.WriteString(style)
	script := `<script>
    $(document).ready(function() {
        var table = $('#diffTable').DataTable({
            "columnDefs": [
                { "orderable": false, "targets": [6,7] }
				// { "visible": false, "targets": [8] }  
            ],
            "pageLength": 50,
            "lengthMenu": [[50, 500, -1], [50, 500, "All"]],
			"order": [[ 1, "desc" ]]
        });
    	$.fn.dataTable.ext.search.push(function(settings, data, dataIndex) {
        	var hideZeros = $('#hideVerse0').prop('checked');
        	if (!hideZeros) return true;
        	return !data[6].endsWith(":0"); 
    	});
    	$('#hideVerse0').prop('checked', true);
    	table.draw();
    	$('#hideVerse0').on('change', function() {
        	table.draw(); 
    	});
    });
</script>
</body>
</html>
`
	_, _ = a.out.WriteString(script)
	_ = a.out.Close()
}

func (a *AlignWriter) minSecFormat(duration float64) string {
	mins := int(duration / 60.0)
	secs := duration - float64(mins)*60.0
	var minStr string
	var delim string
	if int(mins) > 0 {
		minStr = strconv.FormatInt(int64(mins), 10)
		delim = ":"
	}
	secStr := strconv.FormatFloat(secs, 'f', 2, 64)
	return minStr + delim + secStr
}