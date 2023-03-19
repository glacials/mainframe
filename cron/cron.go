package cron

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/robfig/cron/v3"
	"twos.dev/mainframe/calendar"
	"twos.dev/mainframe/dyndns"
	"twos.dev/mainframe/selfupdate"
	"twos.dev/mainframe/speedtest"
)

const (
	production  environment = "production"
	development environment = "development"

	minutely = "@every 1m"
	hourly   = "@every 1h"
	never    = "0 0 5 31 2 ?" // Feb 31 ;)
)

type environment string
type cronSpec struct {
	name      string
	f         func(*log.Logger, string, *sql.DB, *http.ServeMux, *http.Client) error
	intervals map[environment]string
	enabled   bool
}

// Cron fields are, in order:
// second minute hour day-of-month month day-of-week
var (
	// 3am reserved for supervisor.sh to boot me back up if I updated
	crons = []cronSpec{
		{
			name: "calendar",
			f:    calendar.Run,
			intervals: map[environment]string{
				development: minutely,
				production:  "0 0 * * *",
			},
			enabled: false,
		},
		{
			name: "dyndns",
			f:    dyndns.Run,
			intervals: map[environment]string{
				development: never,
				production:  "0 * * * *",
			},
			enabled: true,
		},
		{
			name: "selfupdate",
			f:    selfupdate.Run,
			intervals: map[environment]string{
				development: minutely,
				production:  "0 2 * * *",
			},
			enabled: true,
		},
		{
			name: "speedtest",
			f:    speedtest.Run,
			intervals: map[environment]string{
				development: never,
				production:  "0 5 * * *",
			},
			enabled: true,
		},
	}
)

// Start kicks off all various jobs that should be run occasionally.
func Start(
	logger *log.Logger,
	db *sql.DB,
	version string,
	mux *http.ServeMux,
	gcpClient *http.Client,
) error {
	logger = log.New(logger.Writer(), "[cron] ", logger.Flags())
	environment := development
	if version != "development" {
		environment = production

	}
	c := cron.New()

	if environment == development {
		logger.Println(
			"In development mode; running crons more often & immediately",
		)
		for _, cronDef := range crons {
			f := cronDef.f
			name := cronDef.name
			go func() {
				if err := f(logger, version, db, mux, gcpClient); err != nil {
					logger.Printf("%s failed: %v", name, err)
				}
			}()
		}
	}

	logger.Println("All crons registered")
	c.Start()

	return nil
}
