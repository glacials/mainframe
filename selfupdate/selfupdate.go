package selfupdate

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"syscall"

	"github.com/inconshreveable/go-update"
)

var (
	versionURL  = "https://github.com/glacials/mainframe/releases/latest"
	artifactURL = "https://github.com/glacials/mainframe/releases/download/%s/%s"
	tarfile     = "mainframe-%s-%s-%s.tar.gz"
	version     = "development"
)

type gitHubVersionResponse struct {
	TagName string `json:"tag_name"`
}

// Run runs a self-update if needed.
func Run(logger *log.Logger, version string) error {
	logger = log.New(logger.Writer(), "[selfupdate] ", logger.Flags())

	if version == "development" {
		// logger.Printf("skipping self-update in development mode")
		return nil
	}

	logger.Println("Checking for latest version")

	latestVersion, err := fetchLatestVersion(logger)
	if err != nil {
		return fmt.Errorf("can't fetch latest version: %v", err)
	}

	if latestVersion == version {
		logger.Println("Already running latest, goodbye")
	}

	logger.Printf("Found %v, running %v; updating", latestVersion, version)

	url := fmt.Sprintf(artifactURL, latestVersion, fmt.Sprintf(tarfile, latestVersion, runtime.GOOS, runtime.GOARCH))
	logger.Printf("Downloading %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("can't fetch latest version from GitHub: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code %d downloading artifact from %s", resp.StatusCode, url)
	}

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("can't decompress new version: %v", err)
	}

	tr := tar.NewReader(gzr)
	_, err = tr.Next()
	if err != nil {
		return fmt.Errorf("can't extract new version: %v", err)
	}

	if err := update.Apply(tr, update.Options{}); err != nil {
		return fmt.Errorf("can't update myself: %v", err)
	}

	logger.Println("Finished update, rebooting")

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("can't find my own binary: %v", err)
	}

	if err = syscall.Exec(executable, os.Args, os.Environ()); err != nil {
		return fmt.Errorf("can't reboot myself: %v", err)
	}

	os.Exit(0)

	return nil
}

func fetchLatestVersion(logger *log.Logger) (string, error) {
	client := http.Client{}

	req, err := http.NewRequest("GET", versionURL, nil)
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
