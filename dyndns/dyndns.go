package dyndns

import (
	"fmt"
	"log"
	"os"

	"github.com/jayschwa/go-dyndns"
)

const domain = "mainframe.bifrost.house"

var (
	dyndnsServer   = os.Getenv("DYNDNS_SERVER")
	dyndnsUsername = os.Getenv("DYNDNS_USERNAME")
	dyndnsPassword = os.Getenv("DYNDNS_PASSWORD")
)

// Run updates Google Domains with our current IP.
func Run(logger *log.Logger) error {
	logger = log.New(logger.Writer(), "[dyndns] ", logger.Flags())

	client := dyndns.Service{
		URL:      dyndnsServer,
		Username: dyndnsUsername,
		Password: dyndnsPassword,
	}

	ip, err := client.Update(domain, nil)
	if err != nil {
		return fmt.Errorf("can't update dyndns for %s@%s to %s: %v", dyndnsUsername, dyndnsServer, domain, err)
	}

	logger.Printf("Set %s to %s via %s", domain, ip, dyndnsServer)

	return nil
}
