package fetch

import (
	"context"
	"dataset"
	"dataset/db"
	log "dataset/logger"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type APIDownloadClient struct {
	ctx     context.Context
	bibleId string
}

func NewAPIDownloadClient(ctx context.Context, bibleId string) APIDownloadClient {
	var d APIDownloadClient
	d.ctx = ctx
	d.bibleId = bibleId
	return d
}

func (d *APIDownloadClient) Download(info BibleInfoType) dataset.Status {
	var status dataset.Status
	var directory = filepath.Join(os.Getenv(`FCBH_DATASET_FILES`), info.BibleId)
	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			return log.Error(d.ctx, 500, err, `Could not create directory to store downloaded files.`)
		}
	}
	var download []FilesetType
	if info.AudioOTFileset.Id != `` {
		download = append(download, info.AudioOTFileset)
	}
	if info.AudioNTFileset.Id != `` {
		download = append(download, info.AudioNTFileset)
	}
	if info.TextOTFileset.Id != `` {
		download = append(download, info.TextOTFileset)
	}
	if info.TextNTFileset.Id != `` {
		download = append(download, info.TextNTFileset)
	}
	for _, rec := range download {
		if rec.Type == `text_plain` {
			status = d.downloadPlainText(directory, rec.Id)
			if status.IsErr {
				return status
			}
		} else {
			var locations []LocationRec
			locations, status = d.downloadLocation(rec.Id)
			if status.IsErr {
				if status.Status == 403 {
					locations, status = d.downloadEachLocation(rec)
				} else {
					return status
				}
			}
			if status.IsErr {
				return status
			}
			locations, status = d.sortFileLocations(locations)
			if status.IsErr {
				return status
			}
			directory2 := filepath.Join(directory, rec.Id)
			status = d.downloadFiles(directory2, locations)
			if status.IsErr {
				return status
			}
		}
	}
	return status
}

func (d *APIDownloadClient) downloadPlainText(directory string, filesetId string) dataset.Status {
	var content []byte
	var status dataset.Status
	filename := filesetId + ".json"
	filePath := filepath.Join(directory, filename)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		var get = HOST + "download/" + filesetId + "?v=4&limit=100000"
		fmt.Println("Downloading to", filePath)
		content, status = httpGet(d.ctx, get, false, filesetId)
		if !status.IsErr {
			d.saveFile(filePath, content)
		}
	}
	return status
}

type LocationRec struct {
	BookId   string `json:"book_id"`
	BookName string `json:"book_name"`
	Chapter  int    `json:"chapter_start"`
	Verse    int    `json:"verse_start"`
	URL      string `json:"path"`
	FileSize int    `json:"filesize_in_bytes"`
	Filename string
}
type LocationDownloadRec struct {
	Data []LocationRec `json:"data"`
	Meta any           `json:"meta"`
}

func (d *APIDownloadClient) downloadLocation(filesetId string) ([]LocationRec, dataset.Status) {
	var result []LocationRec
	var status dataset.Status
	var get string
	if strings.Contains(filesetId, `usx`) {
		get = HOST + "bibles/filesets/" + filesetId + "/ALL/1?v=4&limit=100000"
	} else {
		get = HOST + "download/" + filesetId + "?v=4"
	}
	var content []byte
	content, status = httpGet(d.ctx, get, true, filesetId)
	if status.IsErr {
		return result, status
	}
	var response LocationDownloadRec
	err := json.Unmarshal(content, &response)
	if err != nil {
		status = log.Error(d.ctx, 500, err, "Error parsing json for", filesetId)
	} else {
		result = response.Data
	}
	return result, status
}

// downloadEachLocation is used when downloadLocation fails on a 403 error.
// It accesses the location of one chapter at a time using the /bibles/fileset path
func (d *APIDownloadClient) downloadEachLocation(fileset FilesetType) ([]LocationRec, dataset.Status) {
	var result []LocationRec
	var status dataset.Status
	var books []string
	if fileset.Size == `OT` || fileset.Size == `C` {
		books = append(books, db.BookOT...)
	}
	if fileset.Size == `NT` || fileset.Size == `C` {
		books = append(books, db.BookNT...)
	}
	for _, book := range books {
		maxChapter, _ := db.BookChapterMap[book]
		for ch := 1; ch <= maxChapter; ch++ {
			chapter := strconv.Itoa(ch)
			get := HOST + `bibles/filesets/` + fileset.Id + `/` + book + `/` + chapter + `?v=4&`
			var content []byte
			content, status = httpGet(d.ctx, get, false, fileset.Id)
			if status.IsErr {
				return result, status
			}
			var response LocationDownloadRec
			err := json.Unmarshal(content, &response)
			if err != nil {
				status = log.Error(d.ctx, 500, err, "Error parsing json for", fileset.Id)
				return result, status
			}
			for _, data := range response.Data {
				result = append(result, data)
			}
		}
	}
	return result, status
}

func (d *APIDownloadClient) sortFileLocations(locations []LocationRec) ([]LocationRec, dataset.Status) {
	var status dataset.Status
	for i, loc := range locations {
		get, err := url.Parse(loc.URL)
		if err != nil {
			status = log.Error(d.ctx, 500, err, "Could not parse URL", loc.URL)
			if status.IsErr {
				return locations, status
			}
		}
		locations[i].Filename = filepath.Base(get.Path)
	}
	sort.Slice(locations, func(i int, j int) bool {
		return locations[i].Filename < locations[j].Filename
	})
	return locations, status
}

func (d *APIDownloadClient) downloadFiles(directory string, locations []LocationRec) dataset.Status {
	var status dataset.Status
	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			return log.Error(d.ctx, 500, err, "Could not create directory to store downloaded files.")
		}
	}
	for _, loc := range locations {
		filePath := filepath.Join(directory, loc.Filename)
		file, err := os.Stat(filePath)
		if os.IsNotExist(err) || file.Size() != int64(loc.FileSize) {
			fmt.Println("Downloading", loc.Filename)
			var content []byte
			content, status = httpGet(d.ctx, loc.URL, false, loc.Filename)
			if !status.IsErr {
				if len(content) != loc.FileSize {
					log.Warn(d.ctx, "Warning for", loc.Filename, "has an expected size of", loc.FileSize, "but, actual size is", len(content))
				}
				status = d.saveFile(filePath, content)
			}
		}
	}
	return status
}

func (d *APIDownloadClient) saveFile(filePath string, content []byte) dataset.Status {
	var status dataset.Status
	fp, err := os.Create(filePath)
	if err != nil {
		return log.Error(d.ctx, 500, err, "Error Creating file during download.")
	}
	_, err = fp.Write(content)
	if err != nil {
		return log.Error(d.ctx, 500, err, "Error writing to file during download.")
	}
	err = fp.Close()
	if err != nil {
		return log.Error(d.ctx, 500, err, "Error closing file during download.")
	}
	return status
}
