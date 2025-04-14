package align

/*
import (
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/generic"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)
*/
/**
FalsePosFilter is an experimental file to identify false positives.
It used a concept identified by Gordon that if a pattern of errors in a word appears more
than 30 or 40 times it must be a consistent error in the AI process, not an error in the recording.
*/

//* Iterate over verses in alignLine
//* Select words by verse from words table (time this step)
//* Append words to line
//* Append chars as word member
//* Create a map pattern map[string][]Location
//* Type Location struct, book, chapter, verse, word_pos (or ref, word_pos)
//* Create pattern, fa_score, -log10, round to whole num, append num to end of word string.
//* Iterate over all words by book, chapter, verse
//* Skip words that have a score < 1 (const) <try 0 to insure there is little change>
//* Compute pattern, add location to map
//* At end, iterate over map,
//* Append locations over 100 (const) to map[ref+word_pos]word as str cat or struct
//* Print length of map
//* Iterate over line, words
//* Mark each word with location as false pos
//* Change WriteLine to iterate over words and chars, not just chars
//* ignore char errors when in a word that is false positive

/*
func FalsePosFilter(lines []generic.AlignLine) ([]generic.AlignLine, *log.Status) {
	var result []generic.AlignLine
	var status *log.Status

	return result, status
}

func addWords(lines []generic.AlignLine, words []string) ([]generic.AlignLine, *log.Status) {
	var status *log.Status
	for _, line := range lines {
		if len(line.Chars) > 0 {
			line.Chars[0].LineId
		}

	}
	return lines, status
}

func selectWordsByScriptId(scriptId int64) ([]db.Word, *log.Status) {
	query := `SELECT word_id FROM words WHERE script_id = ?`
}

*/
