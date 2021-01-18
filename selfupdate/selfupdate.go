package selfupdate

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var (
	githubVersionURL url.URL = url.URL{Scheme: "https", Host: "github.com", Path: "/glacials/mainframe/releases/latest"}
	version          string  = "development"
)

type gitHubVersionResponse struct {
	TagName string `json:"tag_name"`
}

// Run runs a self-update if needed.
func Run(logger *log.Logger) error {
	logger = log.New(logger.Writer(), "[selfupdate] ", logger.Flags())
	logger.Printf("Starting")

	latestVersion, err := fetchLatestVersion(logger)
	if err != nil {
		return fmt.Errorf("can't fetch latest version: %v", err)
	}

	if latestVersion != version {
		// TODO: Replace myself
	}

	return nil
}

func fetchLatestVersion(logger *log.Logger) (string, error) {
	client := http.Client{}

	req, err := http.NewRequest("GET", githubVersionURL.String(), nil)
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("can't ask GitHub for latest version: %v", err)
	}
	defer resp.Body.Close()

	var body gitHubVersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("can't parse response from GitHub: %v", err)
	}

	return body.TagName, nil
}
