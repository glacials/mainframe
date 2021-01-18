package cron

import (
	"fmt"
	"log"

	"github.com/glacials/mainframe/selfupdate"
	"github.com/glacials/mainframe/speedtest"
	"github.com/robfig/cron"
)

// Start kicks off all various jobs that should be run occasionally.
func Start(logger *log.Logger) error {
	logger = log.New(logger.Writer(), "[cron] ", logger.Flags())
	c := cron.New()

	if err := c.AddFunc("@daily", wrap(selfupdate.Run, logger)); err != nil {
		return fmt.Errorf("Cannot start selfupdate cron: %v", err)
	}
	if err := c.AddFunc("@daily", wrap(speedtest.Run, logger)); err != nil {
		return fmt.Errorf("Cannot start speedtest cron: %v", err)
	}

	return nil
}

func wrap(f func(*log.Logger) error, logger *log.Logger) func() {
	return func() {
		if err := f(logger); err != nil {
			logger.Fatalf("cannot run job: %v", err)
		}
	}
}
