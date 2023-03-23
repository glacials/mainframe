package main

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

const (
	insertIPSQL = `
		INSERT INTO ip_addresses (
			ip_address
		) VALUES (
			$1
		)
	`
	selectIPSQL = `
		SELECT
			ip_address
		FROM
			ip_addresses
		ORDER BY
			created_at DESC
		LIMIT 1
	`
)

var (
	domain         = os.Getenv("DYNDNS_DOMAIN")
	dyndnsServer   = os.Getenv("DYNDNS_SERVER")
	dyndnsUsername = os.Getenv("DYNDNS_USERNAME")
	dyndnsPassword = os.Getenv("DYNDNS_PASSWORD")

	lastKnownPublicIP net.IP = nil
)

// runDyndns updates Google Domains with our current IPu.
func runDynDNS(
	logger *log.Logger,
	version string,
	db *sql.DB,
	_ *http.ServeMux,
	_ *googleClient,
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
	type IP struct {
		Query string
	}

	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return fmt.Errorf("can't get external IP: %w", err)
	}
	defer func() {
		if err := req.Body.Close(); err != nil {
			logger.Printf("can't close request body for getting current IP: %w", err)
		}
	}()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("can't read external IP: %w", err)
	}

	var localIP IP
	json.Unmarshal(body, &localIP)

	localAddr := localIP.Query

	if lastKnownPublicIP == nil {
		if row := db.QueryRow(selectIPSQL); row != nil {
			var ipStr string
			if err := row.Scan(&ipStr); err != nil {
				if err != sql.ErrNoRows {
					return fmt.Errorf("can't bring IP from database to memory: %w", err)
				}
			}
			lastKnownPublicIP = net.ParseIP(ipStr)
		}
	}

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
		if _, err := db.Exec(insertIPSQL, ip.String()); err != nil {
			return fmt.Errorf("can't insert IP into database: %w", err)
		}

		lastKnownPublicIP = ip
		return fmt.Errorf("server says IP is unchanged")
	}
	if err != nil {
		return fmt.Errorf(
			"can't update DNS for %s: %v",
			domain,
			err,
		)
	}

	if _, err := db.Exec(insertIPSQL, ip.String()); err != nil {
		return fmt.Errorf("can't insert IP into database: %w", err)
	}

	lastKnownPublicIP = ip

	logger.Printf("Set %s to %s via %s", domain, ip, dyndnsServer)

	return nil
}
