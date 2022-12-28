package calendar

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

func Run(logger *log.Logger, _ string, db *sql.DB, mux *http.ServeMux, gcpClient *http.Client) error {
	logger = log.New(logger.Writer(), "[calendar] ", logger.Flags())
	ctx := context.Background()

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(gcpClient))
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
	fmt.Println("Upcoming events:")
	if len(events.Items) == 0 {
		return fmt.Errorf("no upcoming events found")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("%v (%v)\n", item.Summary, date)
		}
	}

	return nil
}
