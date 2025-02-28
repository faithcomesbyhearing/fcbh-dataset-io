package input

import (
	"archive/zip"
	"context"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func Glob(ctx context.Context, search string) ([]InputFile, *log.Status) {
	var results []InputFile
	if search != `` {
		files, err := filepath.Glob(search)
		if err != nil {
			return results, log.Error(ctx, 500, err, `Error reading directory`)
		}
		for _, file := range files {
			var input InputFile
			input.Directory = filepath.Dir(file)
			input.Filename = filepath.Base(file)
			results = append(results, input)
		}
	}
	return results, nil
}

func Unzip(ctx context.Context, files []InputFile) ([]InputFile, *log.Status) {
	var results []InputFile
	if len(files) == 1 && filepath.Ext(files[0].Filename) == `.zip` {
		r, err := zip.OpenReader(files[0].FilePath())
		if err != nil {
			return results, log.Error(ctx, 500, err, `Error unzipping file`)
		}
		defer r.Close()
		dest := files[0].Directory
		for _, f := range r.File {
			if f.FileInfo().Name()[0] == '.' {
				continue
			}
			rc, err2 := f.Open()
			if err2 != nil {
				return results, log.Error(ctx, 500, err2, `Error reading zip file`)
			}
			defer rc.Close()
			if f.FileInfo().IsDir() {
				dest = filepath.Join(dest, f.FileInfo().Name())
				os.MkdirAll(dest, f.Mode())
				continue
			}
			path := filepath.Join(dest, f.FileInfo().Name())
			outFile, err3 := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err3 != nil {
				return results, log.Error(ctx, 500, err3, `Error opening file to unzip into`)
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, rc)
			_ = rc.Close()
			if err != nil {
				return results, log.Error(ctx, 500, err, `Error copying file during unzip`)
			}
			results = append(results, InputFile{Filename: f.FileInfo().Name(), Directory: dest})
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].Filename < results[j].Filename
		})
	} else {
		results = files
	}
	return results, nil
}

// SetMediaType function looks at names and sets the Media Type
func SetMediaType(ctx context.Context, file *InputFile) *log.Status {
	fN := file.Filename
	if strings.HasSuffix(fN, `_ET`) || strings.HasSuffix(fN, `_ET.json`) {
		file.MediaType = request.TextPlainEdit
	} else if strings.HasSuffix(fN, `usx`) {
		file.MediaType = request.TextUSXEdit
	} else if (fN[0] == 'A' || fN[0] == 'B') && (fN[1] >= '0' && fN[1] <= '9') {
		file.MediaType = request.Audio
	} else if strings.HasSuffix(fN, `ST.xlsx`) || strings.HasSuffix(fN, `.xlsm`) {
		file.MediaType = request.TextScript
	} else if strings.HasSuffix(fN, `.csv`) {
		file.MediaType = request.TextCSV
	} else if (fN[0] == 'N' || fN[0] == 'O' || fN[0] == 'P') && fN[1] == '1' && fN[2] == '_' {
		file.MediaType = request.Audio
	} else if (fN[0] == 'N' || fN[0] == 'O' || fN[0] == 'P') && fN[1] == '2' && fN[2] == '_' {
		file.MediaType = request.AudioDrama
	} else {
		parts := strings.Split(fN, `_`)
		if len(parts) > 2 {
			ord := parts[1]
			if (ord[0] == 'A' || ord[0] == 'B') && (ord[1] >= '0' && ord[1] <= '9') {
				file.MediaType = request.Audio
			}
		}
	}
	if file.MediaType == `` {
		return log.ErrorNoErr(ctx, 400, `Could not identify media type from filename`)
	}
	return nil
}

func ParseFilenames(ctx context.Context, file *InputFile) *log.Status {
	var status *log.Status
	if file.MediaType == request.TextPlain || file.MediaType == request.TextPlainEdit ||
		file.MediaType == request.TextCSV {
		file.MediaId = strings.Split(file.Filename, `.`)[0]
		test := file.Filename[6]
		if test == 'O' {
			file.Testament = `OT`
		} else if test == 'N' {
			file.Testament = `NT`
		}
		file.FileExt = filepath.Ext(file.Filename)
	} else if file.MediaType == request.TextUSXEdit {
		parts := strings.Split(file.Directory, `/`)
		file.MediaId = parts[len(parts)-1]
		var tmpBookId, tmpBookSeq string
		if len(file.Filename) == 10 {
			tmpBookId = file.Filename[3:6]
			tmpBookSeq = file.Filename[0:3]
		} else if len(file.Filename) == 7 {
			tmpBookId = file.Filename[0:3]
			tmpBookSeq = strconv.Itoa(db.BookSeqMap[tmpBookId])
		} else {
			return log.ErrorNoErr(ctx, 400, `USX files are expected in the format 001GEN.usx or GEN.usx`)
		}
		file.BookId, status = validateBookId(ctx, tmpBookId)
		if status != nil {
			return status
		}
		file.BookSeq = tmpBookSeq
		file.Testament = db.Testament(file.BookId)
		file.FileExt = filepath.Ext(file.Filename)
	} else if file.MediaType == request.TextScript {
		file.MediaId = strings.Split(file.Filename, `.`)[0]
		test := file.Filename[6]
		if test == 'O' {
			file.Testament = `OT`
		} else if test == 'N' {
			file.Testament = `NT`
		}
		file.FileExt = filepath.Ext(file.Filename)
	} else if file.MediaType == request.Audio || file.MediaType == request.AudioDrama {
		fN := file.Filename
		if strings.HasSuffix(fN, `VOX.mp3`) || strings.HasSuffix(fN, `VOX.wav`) {
			status = ParseVOXAudioFilename(ctx, file)
		} else if (fN[0] == 'A' || fN[0] == 'B') && (fN[1] >= '0' && fN[1] <= '9') {
			status = ParseV2AudioFilename(ctx, file)
		} else {
			status = ParseV4AudioFilename(ctx, file)
		}
		if status != nil {
			return status
		}
	} else {
		status = log.ErrorNoErr(ctx, 400, `Type must be one of "text_plain", "text_plain_edit", "text_usx", "audio"`)
	}
	return status
}

func ParseV2AudioFilename(ctx context.Context, file *InputFile) *log.Status {
	var status *log.Status
	var err error
	file.FileExt = filepath.Ext(file.Filename)
	filename := file.Filename[:len(file.Filename)-len(file.FileExt)]
	ab := filename[0]
	if ab == 'A' {
		file.Testament = `OT`
	} else if ab == 'B' {
		file.Testament = `NT`
	}
	seq := filename[1:4]
	file.BookSeq = strings.Trim(seq, `_`)
	file.Chapter, err = strconv.Atoi(file.Filename[6:8])
	if err != nil {
		return log.Error(ctx, 500, err, `Error convert chapter to int`, file.Filename[6:8])
	}
	book := strings.Trim(filename[9:21], `_`)
	file.BookId = db.USFMBookId(ctx, book)
	file.MediaId = filename[21:]
	return status
}

func ParseV4AudioFilename(ctx context.Context, file *InputFile) *log.Status {
	var status *log.Status
	var err error
	file.FileExt = filepath.Ext(file.Filename)
	filename := file.Filename[:len(file.Filename)-len(file.FileExt)]
	filename = strings.Replace(filename, `-`, `_`, -1)
	parts := strings.Split(filename, `_`)
	file.MediaId = parts[0]
	if len(parts) > 1 {
		ab := parts[1][0]
		if ab == 'A' {
			file.Testament = `OT`
		} else if ab == 'B' {
			file.Testament = `NT`
		}
		file.BookSeq = parts[1][1:]
	}
	if len(parts) > 2 {
		file.BookId, status = validateBookId(ctx, parts[2])
		if status != nil {
			return status
		}
	}
	if len(parts) > 3 {
		file.Chapter, err = strconv.Atoi(parts[3])
		if err != nil {
			return log.Error(ctx, 500, err, `Error convert chapter to int`, parts[3])
		}
	}
	if len(parts) > 4 {
		file.Verse = parts[4]
	}
	if len(parts) > 5 {
		file.ChapterEnd, err = strconv.Atoi(parts[5])
		if err != nil {
			return log.Error(ctx, 500, err, `Error convert chapterEnd to int`, parts[5])
		}
	}
	if len(parts) > 6 {
		file.VerseEnd = parts[6]
	}
	return status
}

func ParseVOXAudioFilename(ctx context.Context, file *InputFile) *log.Status {
	var status *log.Status
	var err error
	parts := strings.Split(file.Filename, `_`)
	if len(parts) != 7 {
		return log.ErrorNoErr(ctx, 500, "A VOX. filename is expected to have 7 parts", parts)
	}
	if parts[0][0] == 'N' {
		file.Testament = `NT`
	} else if parts[0][0] == 'O' {
		file.Testament = `OT`
	} else if parts[0][0] == 'P' {
		//file.Testament = `PT` // what should this really be
		file.Testament = db.Testament(file.BookId)
	} else {
		return log.ErrorNoErr(ctx, 500, "Unknown media type", parts[0])
	}
	drama := parts[0]
	langCode := parts[1]
	versionCode := parts[2]
	file.BookSeq = parts[3]
	file.BookId, status = validateBookId(ctx, parts[4])
	if status != nil {
		return status
	}
	file.Chapter, err = strconv.Atoi(parts[5])
	if err != nil {
		return log.Error(ctx, 500, err, `Error convert chapter to int`, parts[5])
	}
	file.MediaId = langCode + versionCode + drama + "DA"
	return status
}

func PruneBooksByRequest(files []InputFile, testament request.Testament) []InputFile {
	var results []InputFile
	for _, f := range files {
		if testament.Has(f.Testament, f.BookId) || f.BookId == `` {
			results = append(results, f)
		}
	}
	return results
}

func UpdateIdent(conn db.DBAdapter, ident *db.Ident, textFiles []InputFile, audioFiles []InputFile) *log.Status {
	var status *log.Status
	_ = updateIdentText(ident, textFiles)
	_ = updateIdentAudio(ident, audioFiles)
	status = conn.InsertReplaceIdent(*ident)
	return status
}

func updateIdentText(ident *db.Ident, files []InputFile) bool {
	var result = false
	if len(files) > 0 {
		f := files[0]
		if f.MediaType != request.TextNone {
			ident.TextSource = f.MediaType
		}
		result = true
		if f.Testament == `OT` {
			ident.TextOTId = f.MediaId
		} else if f.Testament == `NT` {
			ident.TextNTId = f.MediaId
		}
	}
	return result
}

func updateIdentAudio(ident *db.Ident, files []InputFile) bool {
	var result = false
	for _, f := range files {
		result = true
		if f.Testament == `OT` {
			ident.AudioOTId = f.MediaId
		} else if f.Testament == `NT` {
			ident.AudioNTId = f.MediaId
		}
	}
	return result
}

var corrections = map[string]string{
	"PSM": "PSA", // Psalms
	"PRV": "PRO", // Proverbs
	"SOS": "SNG", // Song of Solomon
	"EZE": "EZK", // Ezekiel
	"JOE": "JOL", // Joel
	"NAH": "NAM", // Nahum
	"MRC": "MRK", // Mark
	"LUC": "LUK", // Luke
	"JUA": "JHN", // John
	"HEC": "ACT", // Acts
	"EFE": "EPH", // Ephesians
	"FHP": "PHP", // Philippians
	"1TE": "1TH", // 1 Thessalonians
	"2TE": "2TH", // 2 Thessalonians
	"TTO": "TIT", // Titus
	"TTL": "TIT", // Titus
	"TTS": "TIT", // Titus
	"FHM": "PHM", // Philemon
	"JMS": "JAS", // James
	"SNT": "JAS", // James
	"APO": "REV", // Revelation
}

func validateBookId(ctx context.Context, bookId string) (string, *log.Status) {
	corrected, found := corrections[bookId]
	if found {
		bookId = corrected
	}
	_, ok := db.BookChapterMap[bookId]
	if !ok {
		return bookId, log.ErrorNoErr(ctx, 500, "BookId", bookId, "is not known. Corrections:", corrections)
	}
	return bookId, nil
}

// Parse DBP4 Audio names
//{mediaid}_{A/B}{ordering}_{USFM book code}_{chapter start}[_{verse start}-{chapter stop}_{verse stop}].mp3|webm
//eg: ENGESVN2DA_B001_MAT_001.mp3  (full chapter)
//eg: IRUNLCP1DA_B013_1TH_001_001-001_010.mp3  (partial chapter, verses 1-10)
