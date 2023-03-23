package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// runCalendar fetches calendar events from Google calendar and outputs a few
// upcoming ones.
//
// This was originally going to be a calendar blocker-outer to allow personal
// events to block out work calendars, but my need for that feature was removed
// by other means. This is how far I'd gotten at the time, so I figured I'd keep
// the progress in case my other solution goes away.
func runCalendar(logger *log.Logger, _ string, db *sql.DB, mux *http.ServeMux, google *googleClient) error {
	logger = log.New(logger.Writer(), "[calendar] ", logger.Flags())
	ctx := context.Background()

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(google.http))
	if err != nil {
		return fmt.Errorf("unable to retrieve calendar client: %v", err)
	}

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		return fmt.Errorf(
			"unable to retrieve next ten of the user's events: %v",
			err,
		)
	}
	logger.Println("Upcoming events:")
	if len(events.Items) == 0 {
		return fmt.Errorf("  (none)")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			logger.Printf("  %v (%v)\n", item.Summary, date)
		}
	}

	return nil
}
