package dyndns

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/jayschwa/go-dyndns"
)

var (
	domain         = os.Getenv("DYNDNS_DOMAIN")
	dyndnsServer   = os.Getenv("DYNDNS_SERVER")
	dyndnsUsername = os.Getenv("DYNDNS_USERNAME")
	dyndnsPassword = os.Getenv("DYNDNS_PASSWORD")
)

var (
	lastKnownPublicIP net.IP = nil
)

// Run updates Google Domains with our current IPu.
func Run(
	logger *log.Logger,
	version string,
	_ *sql.DB,
	_ *http.ServeMux,
	_ *http.Client,
) error {
	logger = log.New(logger.Writer(), "[dyndns] ", logger.Flags())

	if version == "development" {
		logger.Println("In development mode; skipping dyndns update")
		return nil
	}

	var unset []string
	if domain == "" {
		unset = append(unset, "DYNDNS_DOMAIN")
	}
	if dyndnsServer == "" {
		unset = append(unset, "DYNDNS_SERVER")
	}
	if dyndnsUsername == "" {
		unset = append(unset, "DYNDNS_USERNAME")
	}
	if dyndnsPassword == "" {
		unset = append(unset, "DYNDNS_PASSWORD")
	}
	if len(unset) > 0 {
		return fmt.Errorf(
			"environment variables %s must be set",
			strings.Join(unset, ", "),
		)
	}

	// We're not really allowed to update using dyndns if our IP address hasn't
	// changed, so we need to keep track of it and check our current one before
	// actually updating Google Domains via dyndns
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return fmt.Errorf("can't check current IP: %v", err)
	}
	defer conn.Close()

	type IP struct {
		Query string
	}

	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return fmt.Errorf("can't get external IP: %v", err)
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("can't read external IP: %v", err)
	}

	var localIP IP
	json.Unmarshal(body, &localIP)

	localAddr := localIP.Query
	if localAddr == lastKnownPublicIP.String() {
		// logger.Printf("IP hasn't changed from %s, so not updating dyndns", lastKnownPublicIP)
		return nil
	}

	logger.Printf(
		"current IP: %s; DNS IP: %s",
		localAddr,
		lastKnownPublicIP.String(),
	)

	client := dyndns.Service{
		URL:      dyndnsServer,
		Username: dyndnsUsername,
		Password: dyndnsPassword,
	}

	ip, err := client.Update(domain, nil)
	if err == dyndns.NoChange {
		lastKnownPublicIP = ip
		return fmt.Errorf("server says IP is unchanged")
	}
	if err != nil {
		return fmt.Errorf(
			"can't update dyndns for %s@%s to %s: %v",
			dyndnsUsername,
			dyndnsServer,
			domain,
			err,
		)
	}

	lastKnownPublicIP = ip

	logger.Printf("Set %s to %s via %s", domain, ip, dyndnsServer)

	return nil
}
