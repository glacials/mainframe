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
	googleoauth "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

const (
	sqlInsertGoogleLink = `
		INSERT INTO google_links (
			access_token,
			token_type,
			refresh_token,
			expires_at
		) VALUES (
			$1, $2, $3, $4
		)
	`
	sqlInsertGoogleLinkScope = `
		INSERT INTO google_link_scopes (
			google_link_id,
			scope
		) VALUES (
			$1, $2
		)
	`
	// sqlUpsertGoogleUser inserts a new google user if it doesn't exist, otherwise
	// updates the existing one.
	sqlUpsertGoogleUser = `
	  INSERT INTO google_users (
		  user_id,
			google_id,
			email,
			verified_email,
			family_name,
			given_name,
			name,
			picture,
			gender,
			hosted_domain,
			link,
			locale
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`
	sqlInsertUser = `
	  INSERT INTO users (
		) VALUES (
		)
	`
	sqlSelectGoogleLink = `
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
	sqlSelectGoogleLinkScope = `
		SELECT
			google_link_id
		FROM google_link_scopes
		WHERE scope = $1
	`
	sqlSelectGoogleUser = `
	  SELECT
		  id
	  FROM google_users
		WHERE google_id = $1
	`
	// sqlSelectUserByGoogleID selects a user by their Google API ID.
	sqlSelectUserByGoogleID = `
	  SELECT
		  id
	  FROM users
		INNER JOIN google_users ON google_users.user_id = users.id
		WHERE google_users.google_id = $1
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

	mux.HandleFunc("/links/google/out", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", config.AuthCodeURL("state-token", oauth2.AccessTypeOffline))
		w.WriteHeader(http.StatusMovedPermanently)
	})

	mux.HandleFunc("/links/google/in", func(w http.ResponseWriter, r *http.Request) {
		token, err := config.Exchange(context.Background(), r.URL.Query().Get("code"))
		if err != nil {
			logger.Printf("unable to retrieve token from web: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s, err := googleoauth.NewService(context.TODO(), option.WithAPIKey(token.AccessToken))
		if err != nil {
			logger.Printf("unable to create oauth2 service: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		u := googleoauth.NewUserinfoV2Service(s)
		userinfo, err := u.Me.Get().Do()
		if err != nil {
			logger.Printf("unable to get userinfo: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		googleUserRow := db.QueryRowContext(r.Context(), sqlSelectGoogleUser, userinfo.Id)
		var googleUserID string
		if googleUserRow.Scan(googleUserID) != nil {
			logger.Printf("unable to get google user id: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		row := db.QueryRowContext(r.Context(), sqlSelectUserByGoogleID, googleUserID)
		var userID int
		if err := row.Scan(userID); err != nil {
			if err == sql.ErrNoRows {
				logger.Printf("tried to log in as Google user %s, but no user exists", googleUserID)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("no user for that Google account exists; did you mean to register?"))
				return
			}
		}

		if err := upsertGoogleUser(r.Context(), db, userID, userinfo); err != nil {
			logger.Printf("unable to create google user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	return client, nil
}

// getTokenFromCLIUser asks the user at the comment line to visit a URL, which
// will give us a token.
func getTokenFromCLIUser(
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
		sqlInsertGoogleLink,
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
			sqlInsertGoogleLinkScope,
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
		rows, err := db.Query(sqlSelectGoogleLinkScope, scope)
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

	row := db.QueryRow(sqlSelectGoogleLink, linkID)

	token := &oauth2.Token{}
	var expiry string
	if err := row.Scan(
		&token.AccessToken,
		&token.TokenType,
		&token.RefreshToken,
		&expiry,
	); err != nil {
		if err == sql.ErrNoRows {
			token, err = getTokenFromCLIUser(logger, config, mux, db)
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

// upsertGoogleUser creates a new user in the database from a Google API user object if
// it doesn't exist, or updates it if it does. The Google user is identified by the ID
// in the userinfo object, not the Mainframe-level userID. This means one Mainframe user
// can have multiple Google users associated with it.
func upsertGoogleUser(ctx context.Context, db *sql.DB, userID int, userinfo *googleoauth.Userinfo) error {
	_, err := db.ExecContext(
		ctx,
		sqlUpsertGoogleUser,
		userID,
		userinfo.Id,
		userinfo.Email,
		userinfo.VerifiedEmail,
		userinfo.FamilyName,
		userinfo.GivenName,
		userinfo.Name,
		userinfo.Picture,
		userinfo.Gender,
		userinfo.Hd,
		userinfo.Link,
		userinfo.Locale,
	)
	if err != nil {
		return fmt.Errorf("can't insert Google user: %w", err)
	}

	return nil
}
