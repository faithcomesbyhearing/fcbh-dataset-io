package fetch

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

type APIDBPClient struct {
	ctx     context.Context
	bibleId string
}

func NewAPIDBPClient(ctx context.Context, bibleId string) APIDBPClient {
	var d APIDBPClient
	d.ctx = ctx
	d.bibleId = bibleId
	return d
}

type FilesetType struct {
	Id        string `json:"id"`
	Type      string `json:"type"`
	Size      string `json:"size"`
	Codec     string `json:"codec"`
	Container string `json:"container"`
	Bitrate   string `json:"bitrate"`
}
type DbpProdType struct {
	Filesets []FilesetType `json:"dbp-prod"`
}
type BooksType struct {
	BookId    string `json:"book_id"`
	Name      string `json:"name"`
	Chapters  []int  `json:"chapters"`
	BookSeq   string `json:"book_seq"`
	Testament string `json:"testament"`
}
type AlphabetType struct {
	Alphabet string `json:"script"`
}
type BibleInfoType struct {
	BibleId            string       `json:"abbr"`
	LanguageISO        string       `json:"iso"`
	LanguageId         int          `json:"language_id"`
	RolvId             string       `json:"language_rolv_code"`
	LanguageName       string       `json:"language"`
	VersionName        string       `json:"name"`
	Alphabet           AlphabetType `json:"alphabet"` // alphabet.script
	Copyright          string       `json:"mark"`
	Books              []BooksType  `json:"books"`
	DbpProd            DbpProdType  `json:"filesets"`
	VersionCode        string
	AudioOTFileset     FilesetType
	AudioNTFileset     FilesetType
	TextOTPlainFileset FilesetType
	TextNTPlainFileset FilesetType
	TextOTUSXFileset   FilesetType
	TextNTUSXFileset   FilesetType
}
type BibleInfoRespType struct {
	Data BibleInfoType `json:"data"`
}

func (d *APIDBPClient) BibleInfo() (BibleInfoType, *log.Status) {
	var result BibleInfoType
	var status *log.Status
	var get = `https://4.dbt.io/api/bibles/` + d.bibleId + `?v=4`
	var response BibleInfoRespType
	body, status := httpGet(d.ctx, get, false, d.bibleId)
	if status != nil {
		return result, status
	}
	//fmt.Println(string(body))
	err := json.Unmarshal(body, &response)
	if err != nil {
		status := log.Error(d.ctx, 500, err, "Error decoding DBP API /bibles JSON")
		return result, status
	}
	result = response.Data
	result.VersionCode = d.bibleId[3:]
	return result, status
}

func (d *APIDBPClient) FindFilesets(info *BibleInfoType, audio request.BibleBrainAudio,
	text request.BibleBrainText, testament request.Testament) {
	textType := text.TextType()
	codec, bitrate := audio.AudioType()
	if testament.OT || len(testament.OTBooks) > 0 {
		info.TextOTPlainFileset = d.searchPlainText(info, `OT`, textType)
		info.TextOTUSXFileset = d.searchUSXText(info, `OT`, textType)
		info.AudioOTFileset = d.searchAudioWithTypePreference(info, `OT`, codec, bitrate, audio.SetTypeCode)
	}
	if testament.NT || len(testament.NTBooks) > 0 {
		info.TextNTPlainFileset = d.searchPlainText(info, `NT`, textType)
		info.TextNTUSXFileset = d.searchUSXText(info, `NT`, textType)
		info.AudioNTFileset = d.searchAudioWithTypePreference(info, `NT`, codec, bitrate, audio.SetTypeCode)
	}
}

func (d *APIDBPClient) searchPlainText(info *BibleInfoType, size string, textType request.MediaType) FilesetType {
	for _, rec := range info.DbpProd.Filesets {
		if rec.Type == `text_plain` && d.hasSize(rec.Size, size) {
			if textType == request.TextPlain || textType == request.TextPlainEdit {
				return rec
			}
		}
	}
	return FilesetType{}
}

func (d *APIDBPClient) searchUSXText(info *BibleInfoType, size string, textType request.MediaType) FilesetType {
	for _, rec := range info.DbpProd.Filesets {
		if rec.Type == `text_usx` && d.hasSize(rec.Size, size) {
			if textType == request.TextUSXEdit || textType == request.TextPlainEdit {
				return rec
			}
		}
	}
	return FilesetType{}
}

func (d *APIDBPClient) searchAudioWithTypePreference(info *BibleInfoType, size string, codec string, bitrate string, preferredType string) FilesetType {
	// If a specific type is preferred, try that first
	if preferredType != "" {
		result := d.searchAudio(info, size, preferredType, codec, bitrate)
		if result.Id != "" {
			return result
		}
	}

	// Fall back to original logic: audio first, then audio_drama
	result := d.searchAudio(info, size, `audio`, codec, bitrate)
	if result.Id != "" {
		return result
	}

	return d.searchAudio(info, size, `audio_drama`, codec, bitrate)
}

func (d *APIDBPClient) searchAudio(info *BibleInfoType, size string, audioType string, codec string, bitrate string) FilesetType {
	for _, rec := range info.DbpProd.Filesets {
		if rec.Type == audioType {
			recCodec := strings.ToUpper(rec.Codec)
			if recCodec == codec || (recCodec == `MP` && codec == `MP3`) {
				if rec.Bitrate == bitrate || (rec.Bitrate == `3kbps` && bitrate == `64kbps`) {
					if d.hasSize(rec.Size, size) {
						if recCodec == `MP` {
							rec.Codec = `MP3`
						}
						if rec.Bitrate == `3kbps` {
							rec.Bitrate = `64kbps`
						}
						return rec
					}
				}
			}
		}
	}
	return FilesetType{}
}

func (d *APIDBPClient) hasSize(recSize string, size string) bool {
	switch recSize {
	case "C", "NTOTP", "OTNTP", "NTPOTP":
		return true
	case "NT", "NTP":
		return size == "NT"
	case "OT", "OTP":
		return size == "OT"
	default:
		return false
	}
}

func (d *APIDBPClient) UpdateIdent(id db.Ident, info BibleInfoType, req request.Request) (db.Ident, *log.Status) {
	var status *log.Status
	if id.BibleId != `` && id.BibleId != req.BibleId {
		return id, log.ErrorNoErr(d.ctx, 400, "Request.yaml has BibleId:", req.BibleId, ", but dataset has BibleId:", id.BibleId)
	}
	id.BibleId = req.BibleId
	if info.AudioOTFileset.Id != `` {
		id.AudioOTId = info.AudioOTFileset.Id
	}
	if info.AudioNTFileset.Id != `` {
		id.AudioNTId = info.AudioNTFileset.Id
	}
	textType := req.TextData.BibleBrain.TextType()
	if textType == request.TextPlain || textType == request.TextPlainEdit {
		if info.TextOTPlainFileset.Id != `` {
			id.TextOTId = info.TextOTPlainFileset.Id
		}
		if info.TextNTPlainFileset.Id != `` {
			id.TextNTId = info.TextNTPlainFileset.Id
		}
	} else if textType == request.TextUSXEdit {
		if info.TextOTUSXFileset.Id != `` {
			id.TextOTId = info.TextOTUSXFileset.Id
		}
		if info.TextNTUSXFileset.Id != `` {
			id.TextNTId = info.TextNTUSXFileset.Id
		}
	}
	id.LanguageISO = req.LanguageISO
	if info.LanguageISO != `` {
		id.LanguageISO = info.LanguageISO
	}
	if info.VersionCode != `` {
		id.VersionCode = info.VersionCode
	}
	if info.LanguageId != 0 {
		id.LanguageId = info.LanguageId
	}
	if info.RolvId != `` {
		tmp, err := strconv.Atoi(info.RolvId)
		if err == nil {
			id.RolvId = tmp
		}
	}
	if info.Alphabet.Alphabet != `` {
		id.Alphabet = info.Alphabet.Alphabet
	}
	if info.LanguageName != `` {
		id.LanguageName = info.LanguageName
	}
	if info.VersionName != `` {
		id.VersionName = info.VersionName
	}
	if textType != request.TextNone {
		id.TextSource = textType
	}
	return id, status
}
