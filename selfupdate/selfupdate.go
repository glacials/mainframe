package selfupdate

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/inconshreveable/go-update"
)

var (
	versionURL  = url.URL{Scheme: "https", Host: "github.com", Path: "/glacials/mainframe/releases/latest"}
	artifactURL = url.URL{Scheme: "https", Host: "github.com", Path: "/glacials/mainframe/releases/download/%s/%s"}
	tarfile     = "mainframe-%s-linux-arm.tar.gz"
	version     = "development"
)

type gitHubVersionResponse struct {
	TagName string `json:"tag_name"`
}

// Run runs a self-update if needed.
func Run(logger *log.Logger, version string) error {
	logger = log.New(logger.Writer(), "[selfupdate] ", logger.Flags())
	logger.Println("Checking for latest version")

	latestVersion, err := fetchLatestVersion(logger)
	if err != nil {
		return fmt.Errorf("can't fetch latest version: %v", err)
	}

	if latestVersion == version {
		logger.Println("Already running latest, goodbye")
	}

	logger.Printf("Found %v, running %v; updating", latestVersion, version)

	url := fmt.Sprintf(artifactURL.String(), latestVersion, fmt.Sprintf(tarfile, latestVersion))
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("can't fetch latest version from GitHub: %v", err)
	}
	defer resp.Body.Close()

	if err := update.Apply(resp.Body, update.Options{}); err != nil {
		return fmt.Errorf("can't update myself: %v", err)
	}

	logger.Println("Finished update, shutting down")
	logger.Println("Depending on system cron to boot me back up")

	os.Exit(0)

	return nil
}

func fetchLatestVersion(logger *log.Logger) (string, error) {
	client := http.Client{}

	req, err := http.NewRequest("GET", versionURL.String(), nil)
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
