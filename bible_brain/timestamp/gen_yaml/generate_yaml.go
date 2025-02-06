package gen_yaml

import (
	"context"
	"fmt"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/decode_yaml/request"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/fetch"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"os"
	"path/filepath"
	"strings"
)

func FindFilesets() ([]fetch.BibleInfoType, *log.Status) {
	var result []fetch.BibleInfoType
	ctx := context.Background()
	testament := request.Testament{NT: true, OT: true}
	audio := request.BibleBrainAudio{MP3_64: true}
	usxText := request.BibleBrainText{TextUSXEdit: true}
	plainText := request.BibleBrainText{TextPlain: true}
	noAudioCount := 0
	noTextCount := 0
	noTextOrAudioCount := 0
	bibles, status := ReadBibles()
	if status != nil {
		return result, status
	}
	for _, bible := range bibles {
		find := fetch.NewAPIDBPClient(ctx, bible.BibleId)
		find.FindFilesets(&bible, audio, usxText, testament)
		hasAudio := false
		hasText := false
		if bible.AudioNTFileset.Id != "" || bible.AudioOTFileset.Id != "" {
			hasAudio = true
		}
		if bible.TextNTUSXFileset.Id != "" || bible.TextOTUSXFileset.Id != "" {
			hasText = true
		}
		if !hasText {
			find.FindFilesets(&bible, audio, plainText, testament)
			if bible.TextNTPlainFileset.Id != "" || bible.TextOTPlainFileset.Id != "" {
				hasText = true
			}
		}
		if !hasAudio && !hasText {
			DisplayBibleError("No Audio and No Text For: ", bible)
			noTextOrAudioCount++
		} else if !hasAudio {
			DisplayBibleError("No Audio For: ", bible)
			noAudioCount++
		} else if !hasText {
			DisplayBibleError("No Text For: ", bible)
			noTextCount++
		} else {
			result = append(result, bible)
		}
	}
	fmt.Println("Number of bibles:", len(result), "No Text:", noTextCount,
		"No Audio:", noAudioCount, "No Text or Audio:", noTextOrAudioCount)
	return result, status
}

func DisplayBibleError(error string, bible fetch.BibleInfoType) {
	fmt.Println(error, bible.BibleId, bible.LanguageISO, bible.LanguageName)
	for _, fs := range bible.DbpProd.Filesets {
		fmt.Println("\t", fs.Id, fs.Type, fs.Size, fs.Codec, fs.Bitrate)
	}
	fmt.Println()
}

func GenerateYaml(bibles []fetch.BibleInfoType) *log.Status {
	var status *log.Status
	for _, bible := range bibles {
		if bible.AudioNTFileset.Id != "" {
			if bible.TextNTUSXFileset.Id != "" {
				status = GenerateOneYaml(bible, bible.AudioNTFileset, bible.TextNTUSXFileset)
			} else if bible.TextNTPlainFileset.Id != "" {
				status = GenerateOneYaml(bible, bible.AudioNTFileset, bible.TextNTPlainFileset)
			}
			if status != nil {
				return status
			}
		}
		if bible.AudioOTFileset.Id != "" {
			if bible.TextOTUSXFileset.Id != "" {
				status = GenerateOneYaml(bible, bible.AudioOTFileset, bible.TextOTUSXFileset)
			} else if bible.TextOTPlainFileset.Id != "" {
				status = GenerateOneYaml(bible, bible.AudioOTFileset, bible.TextOTPlainFileset)
			}
			if status != nil {
				return status
			}
		}
	}
	return status
}

func GenerateOneYaml(bible fetch.BibleInfoType, audio fetch.FilesetType, text fetch.FilesetType) *log.Status {
	// var status *log.Status
	ctx := context.Background()
	var list []string
	list = write(list, 0, "is_new", "yes")
	list = write(list, 0, "dataset_name", audio.Id+"_TS")
	list = write(list, 0, "bible_id", bible.BibleId)
	list = write(list, 0, "username", "GaryNGriswold")
	list = write(list, 0, "email", "gary@shortsands.com")
	list = write(list, 0, "output", "json") // ??
	list = write(list, 0, "testament", "")
	testament := ReduceSize(audio.Size)
	switch testament {
	case "N":
		list = write(list, 1, "nt", "yes")
	case "O":
		list = write(list, 1, "ot", "yes")
	case "C":
		list = write(list, 1, "nt", "yes")
		list = write(list, 1, "ot", "yes")
	default:
		fmt.Println("Unexpected Size:", bible.BibleId, audio.Id, audio.Size, audio.Type)
	}
	list = write(list, 0, "text_data", "")
	list = write(list, 1, "bible_brain", "")
	if text.Type == "text_usx" {
		list = write(list, 2, "text_usx_edit", "yes")
	} else if text.Type == "text_plain" {
		list = write(list, 2, "text_plain", "yes")
	}
	list = write(list, 0, "audio_data", "")
	list = write(list, 1, "bible_brain", "")
	if audio.Codec == "" {
		if strings.Index(audio.Id, "opus") > -1 {
			audio.Codec = "opus"
		} else {
			audio.Codec = "mp3"
		}
	}
	switch audio.Codec {
	case "mp3", "mp", "MP3":
		if audio.Bitrate == "64kbps" {
			list = write(list, 2, "mp3_64", "yes")
		} else if audio.Bitrate == "16kbps" {
			list = write(list, 2, "mp3_16", "yes")
		} else {
			fmt.Println("Unexpected mp3 bitrate", bible.BibleId, audio.Id, audio.Codec, audio.Bitrate)
		}
	case "opus":
		if audio.Bitrate == "16kbps" {
			list = write(list, 2, "opus", "yes")
		} else {
			fmt.Println("Unexpected opus bitrate", bible.BibleId, audio.Id, audio.Codec, audio.Bitrate)
		}
	default:
		fmt.Println("Unexpected codec", bible.BibleId, audio.Id, audio.Codec, audio.Bitrate)
	}
	list = write(list, 0, "timestamps", "")
	list = write(list, 1, "mms_align", "yes")
	list = write(list, 0, "update_dbp", "")
	var filesetList []string
	var updateList []string
	for _, fileset := range bible.DbpProd.Filesets {
		filesetList = append(filesetList, fileset.Id)
		if fileset.Type == "audio" || fileset.Type == "audio_drama" {
			switch ReduceSize(fileset.Size) {
			case "C":
				updateList = append(updateList, fileset.Id)
			case "N":
				if testament == "N" {
					updateList = append(updateList, fileset.Id)
				}
			case "O":
				if testament == "O" {
					updateList = append(updateList, fileset.Id)
				}
			}
		}
	}
	updateStr := strings.Join(updateList, ",")
	list = write(list, 1, "timestamps", "["+updateStr+"]")
	list = write(list, 0, "###", strings.Join(filesetList, ", "))
	filename := filepath.Join("test_data", audio.Id+"_TS.yaml")
	err := os.WriteFile(filename, []byte(strings.Join(list, "")), 0644)
	if err != nil {
		return log.Error(ctx, 500, err, "Writing yaml file")
	}
	return nil
}

func write(list []string, indent int, name string, value string) []string {
	for i := 0; i < indent; i++ {
		list = append(list, "  ")
	}
	list = append(list, name)
	list = append(list, ": ")
	if len(value) > 0 {
		list = append(list, value)
	}
	list = append(list, "\n")
	return list
}

func ReduceSize(size string) string {
	switch size {
	case "C", "NTOTP", "OTNTP", "NTPOTP":
		return "C"
	case "NT", "NTP":
		return "N"
	case "OT", "OTP":
		return "O"
	default:
		return ""
	}
}
