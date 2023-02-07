package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/golang-collections/collections/set"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

const (
	selectScopeSQL = `
		SELECT
			google_link_id
		FROM google_link_scopes
		WHERE scope = $1
	`
	selectLinkSQL = `
		SELECT
			access_token,
			token_type,
			refresh_token,
			expires_at
		FROM google_links
		WHERE id = $1
		ORDER BY expires_at DESC
		LIMIT 1
	`
	insertLinkSQL = `
		INSERT INTO google_links (
			access_token,
			token_type,
			refresh_token,
			expires_at
		) VALUES (
			$1, $2, $3, $4
		)
	`
	insertLinkScopeSQL = `
		INSERT INTO google_link_scopes (
			google_link_id,
			scope
		) VALUES (
			$1, $2
		)
	`
)

var (
	gcpEnvKey = "GCP_CREDENTIALS_FILE"
	gcpEnvVal = os.Getenv("GCP_CREDENTIALS_FILE")
	scopes    = vision.DefaultAuthScopes()
)

func newGCPClient(logger *log.Logger, db *sql.DB, mux *http.ServeMux) (*http.Client, error) {
	if gcpEnvVal == "" {
		return nil, fmt.Errorf("%s is not set", gcpEnvKey)
	}

	mux.HandleFunc("/links/google/out", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Not yet implemented"))
		w.WriteHeader(200)
	})

	b, err := os.ReadFile(gcpEnvVal)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	client, err := newClient(logger, config, mux, db)
	if err != nil {
		return nil, fmt.Errorf("can't get client: %v", err)
	}

	return client, nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(
	logger *log.Logger,
	config *oauth2.Config,
	mux *http.ServeMux,
	db *sql.DB,
) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	logger.Printf("Go to this link to auth Google: \n%v\n", authURL)

	code := make(chan string)
	mux.HandleFunc(
		"/calendar/callback",
		func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("Loading calendar callback page")

			code <- r.URL.Query().Get("code")

			_, err := w.Write(
				[]byte("Google link succeeded. You can close this window."),
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
	result, err := db.Exec(
		insertLinkSQL,
		token.AccessToken,
		token.TokenType,
		token.RefreshToken,
		token.Expiry.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("can't insert token (type=%s, expiry=%s): %v", token.TokenType, token.Expiry, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("can't get last insert id: %v", err)
	}
	if id <= 0 {
		return nil, fmt.Errorf("last insert id is not positive: %d", id)
	}

	for _, scope := range scopes {
		if _, err := db.Exec(
			insertLinkScopeSQL,
			id,
			scope,
		); err != nil {
			return nil, fmt.Errorf("can't insert token scope: %v", err)
		}
	}
	return token, nil
}

// newClient retrieves a token, saves the token if needed, then returns a client
// that uses it.
func newClient(
	logger *log.Logger,
	config *oauth2.Config,
	mux *http.ServeMux,
	db *sql.DB,
) (*http.Client, error) {
	// Link IDs that will supply all the scopes we need.
	var validLinks *set.Set

	for _, scope := range scopes {
		rows, err := db.Query(selectScopeSQL, scope)
		if err != nil {
			return nil, fmt.Errorf("can't select scope from database: %w", err)
		}

		validLinksForThisScope := set.New() // of int
		for rows.Next() {
			var linkID int
			if err := rows.Scan(&linkID); err != nil {
				return nil, fmt.Errorf("can't scan scope row into struct: %w", err)
			}
			validLinksForThisScope.Insert(linkID)
		}

		if validLinks == nil {
			validLinks = validLinksForThisScope
		} else {
			validLinks = validLinks.Intersection(validLinksForThisScope)
		}
	}

	if validLinks.Len() == 0 {
		return nil, fmt.Errorf("no Google token, or at least none that contains the scopes we need")
	}

	var linkID int
	validLinks.Do(func(item interface{}) {
		if intItem, ok := item.(int); ok {
			linkID = intItem
			return
		}
		panic("expected an int argument when iterating over Google link set")
	})

	row := db.QueryRow(selectLinkSQL, linkID)

	token := &oauth2.Token{}
	var expiry string
	if err := row.Scan(
		&token.AccessToken,
		&token.TokenType,
		&token.RefreshToken,
		&expiry,
	); err != nil {
		if err == sql.ErrNoRows {
			token, err = getTokenFromWeb(logger, config, mux, db)
			if err != nil {
				return nil, fmt.Errorf("can't get token from web: %w", err)
			}
		} else {
			return nil, fmt.Errorf("can't look up Google link token: %w", err)
		}
	} else {
		if token.Expiry, err = time.Parse(time.RFC3339, expiry); err != nil {
			return nil, fmt.Errorf("invalid datetime in Google link expires_at column `%s`: %w", expiry, err)
		}
	}

	return config.Client(context.Background(), token), nil
}
