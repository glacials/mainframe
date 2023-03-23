package main

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
	sqlInsertUser = `
	  INSERT INTO users (
		) VALUES (
		)
	`
	sqlSelectGoogleLinkScope = `
		SELECT
			google_link_id
		FROM google_link_scopes
		WHERE scope = $1
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

const (
	envKeyGCP = "GCP_CREDENTIALS_FILE"
)

type googleClient struct {
	http *http.Client
}

type userID int
type googleUserInternalID int

var (
	gcpCredsFile = os.Getenv(envKeyGCP)
	errNoToken   = fmt.Errorf("no Google token, or at least none that contains the scopes we need")
)

func newGoogleClient(logger *log.Logger, db *sql.DB, mux *http.ServeMux) (*googleClient, error) {
	if gcpCredsFile == "" {
		return nil, fmt.Errorf("%s is not set", envKeyGCP)
	}

	b, err := os.ReadFile(gcpCredsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	loginConfig, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	registerConfig, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	mux.HandleFunc("/login/google", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", loginConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline))
		w.WriteHeader(http.StatusMovedPermanently)
	})

	mux.HandleFunc("/register/google", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", registerConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline))
		w.WriteHeader(http.StatusMovedPermanently)
	})

	mux.HandleFunc("/login/google/callback", func(w http.ResponseWriter, r *http.Request) {
		token, err := loginConfig.Exchange(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			logger.Printf("unable to retrieve token from web: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		service, err := googleoauth.NewService(r.Context(), option.WithAPIKey(token.AccessToken))
		if err != nil {
			logger.Printf("unable to create oauth2 service: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userinfo, err := googleoauth.NewUserinfoV2Service(service).Me.Get().Do()
		if err != nil {
			logger.Printf("unable to get userinfo: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		googleUserRow := db.QueryRowContext(
			r.Context(),
			`
	      SELECT id
        FROM google_users
        WHERE google_id = $1
	    `,
			userinfo.Id,
		)
		var id googleUserInternalID
		if googleUserRow.Scan(id) != nil {
			logger.Printf("unable to get google user id: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		row := db.QueryRowContext(r.Context(), sqlSelectUserByGoogleID, id)
		var userID int
		if err := row.Scan(userID); err != nil {
			if err == sql.ErrNoRows {
				logger.Printf("tried to log in as Google user %s, but no attached user exists", id)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("no user for that Google account exists; <a href='/register/google'>register</a>?"))
				return
			}
		}

		if err := upsertGoogleUser(r.Context(), db, userID, userinfo); err != nil {
			logger.Printf("unable to create google user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/register/google/callback", func(w http.ResponseWriter, r *http.Request) {
		token, err := loginConfig.Exchange(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			logger.Printf("unable to retrieve token from web: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		service, err := googleoauth.NewService(r.Context(), option.WithAPIKey(token.AccessToken))
		if err != nil {
			logger.Printf("unable to create oauth2 service: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userinfo, err := googleoauth.NewUserinfoV2Service(service).Me.Get().Do()
		if err != nil {
			logger.Printf("unable to get userinfo: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = db.ExecContext(
			r.Context(),
			`
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
		`,
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
	})

	return &googleClient{
		http: &http.Client{},
	}, nil
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

// tokenForUser returns a Google API token that fulfills the given scope for the
// given Google user internal ID. If there is none, errNoToken is returned.
func tokenForUser(
	ctx context.Context,
	logger *log.Logger,
	mux *http.ServeMux,
	db *sql.DB,
	googleUserInternalID googleUserInternalID,
	scope string,
) (*oauth2.Token, error) {
	row := db.QueryRowContext(
		ctx,
		`
		SELECT
			access_token,
			token_type,
			refresh_token,
			expires_at
		FROM google_links
		INNER JOIN google_link_scopes
			ON google_links.id = google_link_scopes.google_link_id
		INNER JOIN google_users
			ON google_links.google_user_id = google_users.id
		WHERE
			google_user.id = $1
		AND
			google_link_scopes.scope = $2
		ORDER BY expires_at DESC
		LIMIT 1
	  `,
		googleUserInternalID,
		scope,
	)
	token := oauth2.Token{}
	var expiry string
	if err := row.Scan(
		&token.AccessToken,
		&token.TokenType,
		&token.RefreshToken,
		&expiry,
	); err != nil {
		return nil, fmt.Errorf("can't look up Google link token: %w", err)
	}
	token.Expiry, err = time.Parse(time.RFC3339, expiry)
	if err != nil {
		return nil, fmt.Errorf("invalid datetime in Google link expires_at column `%s`: %w", expiry, err)
	}

	return &token, nil
}

// upsertGoogleUser creates a new user in the database from a Google API user object if
// it doesn't exist, or updates it if it does. The Google user is identified by the ID
// in the userinfo object, not the Mainframe-level userID. This means one Mainframe user
// can have multiple Google users associated with it.
func upsertGoogleUser(ctx context.Context, db *sql.DB, userID int, userinfo *googleoauth.Userinfo) error {

	if err != nil {
		return fmt.Errorf("can't insert Google user: %w", err)
	}

	return nil
}
