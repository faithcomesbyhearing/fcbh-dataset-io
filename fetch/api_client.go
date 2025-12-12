package fetch

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/faithcomesbyhearing/fcbh-dataset-io/logger"
)

// getDBPHost returns the DBP base host from environment variable or default
func getDBPHost() string {
	host := os.Getenv("DBP_API_HOST")
	if host == "" {
		return "https://4.dbt.io"
	}
	// Ensure host has protocol
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	// Remove trailing slashes
	host = strings.TrimRight(host, "/")
	return host
}

func HttpGet(ctx context.Context, url string, desc string) ([]byte, *log.Status) {
	return httpGet(ctx, url, false, desc)
}

func httpGet(ctx context.Context, url string, ok403 bool, desc string) ([]byte, *log.Status) {
	var body []byte
	host := getDBPHost() + "/api/"
	if strings.Contains(url, host) {
		url += `&limit=100000&key=` + os.Getenv(`FCBH_DBP_KEY`)
	}
	resp, err := http.Get(url)
	if err != nil {
		return body, log.Error(ctx, 0, err, "Error in DBP API request for:", desc)
	}
	defer resp.Body.Close()
	if ok403 && resp.StatusCode == 403 {
		var status log.Status
		status.Status = 403
		return body, &status
	}
	if resp.Status[0] != '2' {
		return body, log.ErrorNoErr(ctx, resp.StatusCode, resp.Status, desc)
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return body, log.Error(ctx, resp.StatusCode, err, "Error reading DBP API response for:", desc)
	}
	return body, nil
}
