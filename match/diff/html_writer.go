package diff

import (
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"github.com/sergi/go-diff/diffmatchpatch"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type HTMLWriter struct {
	ctx         context.Context
	datasetName string
	diffMatch   *diffmatchpatch.DiffMatchPatch
	out         *os.File
	diffCount   int
	insertSum   int
	deleteSum   int
}

func NewHTMLWriter(ctx context.Context, datasetName string) HTMLWriter {
	var h HTMLWriter
	h.ctx = ctx
	h.datasetName = datasetName
	h.diffMatch = diffmatchpatch.New()
	return h
}

func (h *HTMLWriter) WriteReport(baseDataset string, records []Pair, languageISO string, fileMap string,
	asr request.SpeechToText) (string, *log.Status) {
	var err error
	var model string
	switch asr {
	case request.SpeechToText{MMS: true}:
		model = "Model: MMS"
	case request.SpeechToText{MMSAdapter: true}:
		model = "Model: MMS Adapter"
	case request.SpeechToText{Wav2Vec2ASR: true}:
		model = "Model: WAV2Vec2 Word"
	default:
		model = ""
	}
	h.out, err = os.Create(filepath.Join(os.Getenv(`FCBH_DATASET_TMP`), h.datasetName+"_compare.html"))
	if err != nil {
		return "", log.Error(h.ctx, 500, err, `Error creating output file for diff`)
	}
	filename := h.WriteHeading(baseDataset, languageISO, model)
	for _, pair := range records {
		h.WriteLine(pair)
	}
	h.WriteEnd(fileMap)
	return filename, nil
}

func (h *HTMLWriter) WriteHeading(baseDataset string, languageISO string, model string) string {
	head := `<!DOCTYPE html>
<html>
 <head>
  <meta charset="utf-8">
  <title>File Difference</title>
`
	_, _ = h.out.WriteString(head)
	_, _ = h.out.WriteString(`<link rel="stylesheet" type="text/css" href="https://cdn.datatables.net/1.10.21/css/jquery.dataTables.css">`)
	_, _ = h.out.WriteString("</head><body>\n")
	_, _ = h.out.WriteString(`<h2 style="text-align:center">Compare `)
	_, _ = h.out.WriteString(baseDataset)
	_, _ = h.out.WriteString(` to `)
	_, _ = h.out.WriteString(h.datasetName)
	_, _ = h.out.WriteString("</h2>\n")
	_, _ = h.out.WriteString(`<h3 style="text-align:center">`)
	_, _ = h.out.WriteString(model)
	_, _ = h.out.WriteString(`   ASR ISO `)
	_, _ = h.out.WriteString(languageISO)
	_, _ = h.out.WriteString(`</h3>`)
	_, _ = h.out.WriteString(`<h3 style="text-align:center">`)
	loc, _ := time.LoadLocation("America/Denver")
	_, _ = h.out.WriteString(time.Now().In(loc).Format(`Mon Jan 2 2006 03:04:05 pm MST`))
	_, _ = h.out.WriteString("</h3>\n")
	_, _ = h.out.WriteString(`<h3 style="text-align:center">RED characters are those in `)
	_, _ = h.out.WriteString(baseDataset)
	_, _ = h.out.WriteString(` only, while GREEN characters are in `)
	_, _ = h.out.WriteString(h.datasetName)
	_, _ = h.out.WriteString(" only</h3>\n")
	checkbox := `<div style="text-align: center; margin: 10px;">
		<input type="checkbox" id="hideVerse0" checked><label for="hideVerse0">Hide Headings</label>
	</div>
`
	_, _ = h.out.WriteString(checkbox)
	directoryInput := `<div style="text-align: center; margin: 10px;">
		<label for="directory">Directory of Audio Files: </label><input type="text" id="directory" size="100" value="./">
	</div>`

	_, _ = h.out.WriteString(directoryInput)
	_, _ = h.out.WriteString("<audio id='validateAudio'></audio>\n")
	table := `<table id="diffTable" class="display">
    <thead>
    <tr>
        <th>Line</th>
		<th>Len</th>
		<th>Button</th>
        <th>Ref</th>
		<th>Text Comparison</th>
    </tr>
    </thead>
    <tbody>
`
	_, _ = h.out.WriteString(table)
	return h.out.Name()
}

func (h *HTMLWriter) WriteLine(verse Pair) {
	largest := verse.LargestLength()
	if largest > 2 {
		h.diffCount++
		inserts := verse.Inserts()
		h.insertSum += inserts
		deletes := verse.Deletes()
		h.deleteSum += deletes
		//errPct := verse.ErrorPct(inserts, deletes)
		_, _ = h.out.WriteString("<tr>\n")
		if verse.ScriptNum != "" {
			h.writeCell(verse.ScriptNum)
		} else {
			h.writeCell(strconv.Itoa(verse.ScriptId()))
		}
		//h.writeCell(strconv.FormatFloat(errPct, 'f', 0, 64))
		//h.writeCell(strconv.Itoa(inserts + deletes))
		//h.writeCell(strconv.Itoa(int(math.Abs(float64(inserts - deletes)))))
		h.writeCell(strconv.Itoa(largest))
		//h.writeCell(h.minSecFormat(verse.beginTS))
		var params []string
		params = append(params, "this")
		params = append(params, "'"+verse.Ref.BookId+"'")
		params = append(params, strconv.Itoa(verse.Ref.ChapterNum))
		params = append(params, strconv.FormatFloat(verse.BeginTS, 'f', 4, 64))
		params = append(params, strconv.FormatFloat(verse.EndTS, 'f', 4, 64))
		h.writeCell("<button title=\"" + h.minSecFormat(verse.BeginTS) + "\" onclick=\"playVerse(" + strings.Join(params, ",") + ")\">Play</button>")
		//h.writeCell(`+` + strconv.Itoa(inserts) + ` -` + strconv.Itoa(deletes))
		h.writeCell(verse.Ref.Description())
		h.writeCell(verse.HTML)
		_, _ = h.out.WriteString("</tr>\n")
	}
}

func (h *HTMLWriter) writeCell(content string) {
	_, _ = h.out.WriteString(`<td>`)
	_, _ = h.out.WriteString(content)
	_, _ = h.out.WriteString(`</td>`)
}

func (h *HTMLWriter) WriteEnd(filenameMap string) {
	table := `</tbody>
	</table>
`
	_, _ = h.out.WriteString(table)
	_, _ = h.out.WriteString(`<p>Total Inserted Chars `)
	_, _ = h.out.WriteString(strconv.Itoa(h.insertSum))
	_, _ = h.out.WriteString(`, Total Deleted Chars `)
	_, _ = h.out.WriteString(strconv.Itoa(h.deleteSum))
	_, _ = h.out.WriteString("</p>\n")
	_, _ = h.out.WriteString(`<p>`)
	_, _ = h.out.WriteString("Total Difference Count: ")
	_, _ = h.out.WriteString(strconv.Itoa(h.diffCount))
	_, _ = h.out.WriteString("</p>\n")
	_, _ = h.out.WriteString(`<script type="text/javascript" src="https://code.jquery.com/jquery-3.5.1.js"></script>`)
	_, _ = h.out.WriteString("\n")
	_, _ = h.out.WriteString(`<script type="text/javascript" src="https://cdn.datatables.net/1.10.21/js/jquery.dataTables.js"></script>`)
	_, _ = h.out.WriteString("\n")
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
	</style>
`
	_, _ = h.out.WriteString(style)
	script := `<script>
    $(document).ready(function() {
        var table = $('#diffTable').DataTable({
            "columnDefs": [
                { "orderable": false, "targets": [2,4] }
				// { "visible": false, "targets": [8] }  
            ],
            "pageLength": 50,
            "lengthMenu": [[50, 500, -1], [50, 500, "All"]],
			"order": [[ 1, "desc" ]]
        });
    	$.fn.dataTable.ext.search.push(function(settings, data, dataIndex) {
        	var hideZeros = $('#hideVerse0').prop('checked');
        	if (!hideZeros) return true;
        	return !data[3].endsWith(":0"); 
    	});
    	$('#hideVerse0').prop('checked', true);
    	table.draw();
    	$('#hideVerse0').on('change', function() {
        	table.draw(); 
    	});
    });
	function playVerse(button, book, chapter, startTime, endTime) {
`
	_, _ = h.out.WriteString(script)
	_, _ = h.out.WriteString("\t" + filenameMap)
	script = `filename = fileMap[book + chapter]
		let directory = document.getElementById("directory").value
		audioFile = directory + '/' + filename;
		const audio = document.getElementById('validateAudio');
		/* const rect = button.getBoundingClientRect();
    	audio.style.position = 'fixed';
    	audio.style.left = rect.left + window.scrollX + 'px';
    	audio.style.top = (rect.bottom + window.scrollY + 5) + 'px'; */
		audio.src = audioFile;
		audio.playbackRate = 0.75;
		audio.currentTime = startTime;
		audio.controls = false;
		audio.play();
		audio.ontimeupdate = function() {
			if (audio.currentTime >= endTime) {
				audio.pause();
				audio.ontimeupdate = null;
			}
		}
	}
    </script>
</body>
</html>
`
	_, _ = h.out.WriteString(script)
	_ = h.out.Close()
}

func (h *HTMLWriter) minSecFormat(duration float64) string {
	if duration > 0.5 {
		duration -= 0.5
	} else {
		duration = 0.0
	}
	mins := int(duration / 60.0)
	secs := duration - float64(mins)*60.0
	var minStr string
	var delim string
	if int(mins) > 0 {
		minStr = strconv.FormatInt(int64(mins), 10)
		delim = ":"
	}
	secStr := strconv.FormatFloat(secs, 'f', 0, 64)
	return minStr + delim + secStr
}
