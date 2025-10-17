package update

import "database/sql"

type Timestamp struct {
	TimestampId int64
	VerseStr    string
	VerseEnd    sql.NullString // On update set to null
	VerseSeq    int
	BeginTS     float64
	EndTS       float64
	AudioFile   string
}
