package cron

import (
	"fmt"
	"log"

	"github.com/glacials/mainframe/dyndns"
	"github.com/glacials/mainframe/selfupdate"
	"github.com/glacials/mainframe/speedtest"
	"github.com/robfig/cron"
)

var (
	// Cron fields are, in order:
	// second minute hour day-of-month month day-of-week
	selfupdateInterval = "0 0 2 * * *" // Update myself at 3am every day
	// 3am reserved for mainframe_helper.sh to boot me back up if I updated
	dyndnsInterval    = "0 0 4 * * *" // Update DNS at 4am every day
	speedtestInterval = "0 0 5 * * *" // Run a speed test at 4am every day

	minutely = "@every 1m" // For use in development
)

// Start kicks off all various jobs that should be run occasionally.
func Start(logger *log.Logger, version string) error {
	logger = log.New(logger.Writer(), "[cron] ", logger.Flags())

	c := cron.New()

	if version == "development" {
		selfupdateInterval = minutely
		//speedtestInterval = minutely
	}

	if err := c.AddFunc(selfupdateInterval, func() {
		if err := selfupdate.Run(logger, version); err != nil {
			logger.Fatalf("selfupdate failed: %v", err)
		}
	}); err != nil {
		return fmt.Errorf("cannot start selfupdate cron: %v", err)
	}
	logger.Println("Registered selfupdate")

	if err := c.AddFunc(speedtestInterval, func() {
		if err := speedtest.Run(logger); err != nil {
			logger.Fatalf("speedtest failed: %v", err)
		}
	}); err != nil {
		return fmt.Errorf("cannot start speedtest cron: %v", err)
	}
	logger.Println("Registered speedtest")

	if err := c.AddFunc(dyndnsInterval, func() {
		if err := dyndns.Run(logger); err != nil {
			logger.Fatalf("dyndns failed: %v", err)
		}
	}); err != nil {
		return fmt.Errorf("cannot start dyndns cron: %v", err)
	}
	logger.Println("Registered dyndns")

	logger.Println("All crons registered")
	c.Start()

	return nil
}
