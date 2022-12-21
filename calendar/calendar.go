package calendar

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	insertSQL = `
		INSERT INTO google_links (
			access_token,
			token_type,
			refresh_token,
			expires_at,
			scope
		) VALUES (
			$1, $2, $3, $4, $5
		)
	`
	scope     = calendar.CalendarEventsScope
	selectSQL = `
		SELECT
			access_token,
			token_type,
			refresh_token,
			expires_at
		FROM google_links
		ORDER_BY expires_at DESC
		LIMIT 1
	`
)

var GCP_CREDENTIALS_FILE = os.Getenv("GCP_CREDENTIALS_FILE")

// newClient retrieves a token, saves the token if needed, then returns a client
// that uses it.
func newClient(
	logger *log.Logger,
	config *oauth2.Config,
	mux *http.ServeMux,
	db *sql.DB,
) (*http.Client, error) {
	row := db.QueryRow(selectSQL)

	token := &oauth2.Token{}
	if err := row.Scan(
		&token.AccessToken,
		&token.TokenType,
		&token.RefreshToken,
		&token.Expiry,
	); err != nil {
		if err != nil {
			token, err = getTokenFromWeb(logger, config, mux, db)
			if err != nil {
				return nil, fmt.Errorf("can't get token from web: %v", err)
			}
		}
	}

	return config.Client(context.Background(), token), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(
	logger *log.Logger,
	config *oauth2.Config,
	mux *http.ServeMux,
	db *sql.DB,
) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	logger.Printf("Go to this link to auth Google Calendar: \n%v\n", authURL)

	code := make(chan string)
	mux.HandleFunc(
		"/calendar/callback",
		func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("Loading calendar callback page")

			code <- r.URL.Query().Get("code")

			_, err := w.Write(
				[]byte("Google Calendar link succeeded. You can close this window."),
			)
			if err != nil {
				logger.Printf("can't write to response: %v", err)
			}
		},
	)

	token, err := config.Exchange(context.TODO(), <-code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	if _, err := db.Exec(
		insertSQL,
		token.AccessToken,
		token.TokenType,
		token.RefreshToken,
		token.Expiry,
		scope,
	); err != nil {
		return nil, fmt.Errorf("can't insert token: %v", err)
	}
	return token, nil
}

func Run(logger *log.Logger, _ string, db *sql.DB, mux *http.ServeMux) error {
	logger = log.New(os.Stdout, "[calendar] ", logger.Flags())

	if GCP_CREDENTIALS_FILE == "" {
		return fmt.Errorf("GCP_CREDENTIALS_FILE is not set")
	}

	mux.HandleFunc("/calendar", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Not yet implemented"))
		w.WriteHeader(200)
	})

	ctx := context.Background()
	b, err := os.ReadFile(GCP_CREDENTIALS_FILE)
	if err != nil {
		return fmt.Errorf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		return fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}

	client, err := newClient(logger, config, mux, db)
	if err != nil {
		return fmt.Errorf("can't get client: %v", err)
	}

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("Unable to retrieve Calendar client: %v", err)
	}

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		return fmt.Errorf(
			"Unable to retrieve next ten of the user's events: %v",
			err,
		)
	}
	fmt.Println("Upcoming events:")
	if len(events.Items) == 0 {
		return fmt.Errorf("No upcoming events found.")
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
