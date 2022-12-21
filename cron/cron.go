package cron

import (
	"database/sql"
	"fmt"
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

	minutely = "@every 2m"
	hourly   = "@every 1h"
)

type environment string
type cronSpec struct {
	name      string
	f         func(*log.Logger, string, *sql.DB, *http.ServeMux) error
	intervals map[environment]string
}

// Cron fields are, in order:
// second minute hour day-of-month month day-of-week
var (
	// 3am reserved for mainframe_helper.sh to boot me back up if I updated
	crons = []cronSpec{
		{
			name: "calendar",
			f:    calendar.Run,
			intervals: map[environment]string{
				development: minutely,
				production:  "0 0 * * *",
			},
		},
		{
			name: "dyndns",
			f:    dyndns.Run,
			intervals: map[environment]string{
				development: minutely,
				production:  "0 * * * *",
			},
		},
		{
			name: "selfupdate",
			f:    selfupdate.Run,
			intervals: map[environment]string{
				development: minutely,
				production:  "0 2 * * *",
			},
		},
		{
			name: "speedtest",
			f:    speedtest.Run,
			intervals: map[environment]string{
				development: hourly,
				production:  "0 5 * * *"},
		},
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
				if err := f(logger, version, db, mux); err != nil {
					logger.Printf("%s failed: %v", name, err)
				}
			}()
		}
	}

	for _, cronSpec := range crons {
		logger.Printf(
			"Registering %s for %s\n",
			cronSpec.name,
			cronSpec.intervals[environment],
		)
		if _, err := c.AddFunc(cronSpec.intervals[environment], func() {
			logger.Printf("Kicking off %s\n", cronSpec.name)
			if err := cronSpec.f(logger, version, db, mux); err != nil {
				logger.Printf("%s failed: %v", cronSpec.name, err)
			}
		}); err != nil {
			return fmt.Errorf("cannot start %s cron: %v", cronSpec.name, err)
		}
		logger.Printf("Registered %s\n", cronSpec.name)
	}

	logger.Println("All crons registered")
	c.Start()

	return nil
}
