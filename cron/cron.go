package cron

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/robfig/cron"
	"twos.dev/mainframe/calendar"
	"twos.dev/mainframe/dyndns"
	"twos.dev/mainframe/selfupdate"
	"twos.dev/mainframe/speedtest"
)

type cronDefinition struct {
	name     string
	f        func(*log.Logger, string, *sql.DB, *http.ServeMux) error
	interval string
}

// Cron fields are, in order:
// second minute hour day-of-month month day-of-week
var (
	minutely = "@every 1m" // For use in development

	// 3am reserved for mainframe_helper.sh to boot me back up if I updated
	funcs = []cronDefinition{
		{name: "calendar", f: calendar.Run, interval: "0 0 0 * * *"},
		{name: "dyndns", f: dyndns.Run, interval: "0 * * * * *"},
		{name: "selfupdate", f: selfupdate.Run, interval: "0 0 2 * * *"},
		{name: "speedtest", f: speedtest.Run, interval: "0 0 5 * * *"},
	}
)

// Start kicks off all various jobs that should be run occasionally.
func Start(
	logger *log.Logger,
	db *sql.DB,
	version string,
	mux *http.ServeMux,
) error {
	logger = log.New(logger.Writer(), "[cron] ", logger.Flags())

	c := cron.New()

	if version == "development" {
		logger.Println("In development mode; running now and using 1m intervals")
		for _, cronDef := range funcs {
			cronDef.interval = minutely
			f := cronDef.f
			name := cronDef.name
			go func() {
				if err := f(logger, version, db, mux); err != nil {
					logger.Printf("%s failed: %v", name, err)
				}
			}()
		}
	}

	for _, cronDef := range funcs {
		if err := c.AddFunc(cronDef.interval, func() {
			if err := cronDef.f(logger, version, db, mux); err != nil {
				logger.Fatalf("%s failed: %v", cronDef.name, err)
			}
		}); err != nil {
			return fmt.Errorf("cannot start %s cron: %v", cronDef.name, err)
		}
		logger.Printf("Registered %s\n", cronDef.name)
	}

	logger.Println("All crons registered")
	c.Start()

	return nil
}
