package gen_yaml

import (
	"context"
	"encoding/json"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/fetch"
	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
	"os"
	"strings"
)

type BiblesResponse struct {
	Data []fetch.BibleInfoType `json:"data"`
	Meta Meta                  `json:"meta"`
}

type Meta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	CurrentPage int    `json:"current_page"`
	FromPage    int    `json:"from"`
	LastPage    int    `json:"last_page"`
	NextPageURL string `json:"next_page_url"`
	PerPage     int    `json:"per_page"`
	PrevPageURL string `json:"prev_page_url"`
	To          int    `json:"to"`
	Total       int    `json:"total"`
}

func FetchBibles() *log.Status {
	var result []fetch.BibleInfoType
	ctx := context.Background()
	url := "https://4.dbt.io/api/bibles?v=4"
	for {
		if url == "" {
			break
		}
		content, status := fetch.HttpGet(ctx, url, "")
		if status != nil {
			return status
		}
		var response BiblesResponse
		err := json.Unmarshal(content, &response)
		if err != nil {
			return log.Error(ctx, 500, err, "Error unmarshalling bibles API query")
		}
		result = append(result, response.Data...)
		url = response.Meta.Pagination.NextPageURL
		url = strings.Replace(url, "?page", "?v=4&page", -1)
	}
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return log.Error(ctx, 500, err, "Error marshalling bibles API query")
	}
	filename := "bible_brain_api.json"
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return log.Error(ctx, 500, err, "Error writing bibles API query")
	}
	return nil
}

func ReadBibles() ([]fetch.BibleInfoType, *log.Status) {
	var bibles []fetch.BibleInfoType
	ctx := context.Background()
	content, err := os.ReadFile("bible_brain_api.json")
	if err != nil {
		return bibles, log.Error(ctx, 500, err, "Reading file")
	}
	err = json.Unmarshal(content, &bibles)
	if err != nil {
		return bibles, log.Error(ctx, 500, err, "Error unmarshalling bibles")
	}
	return bibles, nil
}
