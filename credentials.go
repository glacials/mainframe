package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
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
	baseURL   = "http://localhost:9000"
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

	loginConfig.RedirectURL = baseURL + "/login/google/callback"
	registerConfig.RedirectURL = baseURL + "/register/google/callback"

	mux.HandleFunc("/login/google", func(w http.ResponseWriter, r *http.Request) {
		state, err := generateState(r.Context(), db)
		if err != nil {
			logger.Printf("unable to generate random state: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", loginConfig.AuthCodeURL(state, oauth2.AccessTypeOffline))
		w.WriteHeader(http.StatusMovedPermanently)
	})

	mux.HandleFunc("/register/google", func(w http.ResponseWriter, r *http.Request) {
		state, err := generateState(r.Context(), db)
		if err != nil {
			logger.Printf("unable to generate random state: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", registerConfig.AuthCodeURL(state, oauth2.AccessTypeOffline))
		w.WriteHeader(http.StatusMovedPermanently)
	})

	mux.HandleFunc("/login/google/callback", func(w http.ResponseWriter, r *http.Request) {
		if err := consumeState(r.Context(), db, r.FormValue("state")); err != nil {
			logger.Printf("unable to consume state: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		token, err := loginConfig.Exchange(r.Context(), r.FormValue("code"))
		if err != nil {
			logger.Printf("unable to retrieve register token from web: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
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

		row := db.QueryRowContext(
			r.Context(),
			`
			SELECT
				id
			FROM users
			INNER JOIN google_users ON google_users.user_id = users.id
			WHERE google_users.google_id = $1
			`,
			id,
		)
		var userID int
		if err := row.Scan(userID); err != nil {
			if err == sql.ErrNoRows {
				logger.Printf("tried to log in as Google user %d, but no attached user exists", id)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("no user for that Google account exists; <a href='/register/google'>register</a>?"))
				return
			}
		}
	})

	mux.HandleFunc("/register/google/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := consumeState(r.Context(), db, r.FormValue("state")); err != nil {
			logger.Printf("unable to consume state: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		token, err := registerConfig.Exchange(r.Context(), r.FormValue("code"))
		if err != nil {
			logger.Printf("unable to retrieve login token from web: %v", err)
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
			SELECT id, user_id
			FROM google_users
			WHERE external_id = $2
	    `,
			userinfo.Id,
		)

		var (
			googleUserID int
			userID       int
		)
		if googleUserRow.Scan(&googleUserID, &userID) != sql.ErrNoRows {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(
				`
			  This Google account is already bound to a Mainframe account. <a href='/login/google'>Log in</a>?
			  `,
			))
			return
		}
		if googleUserID != 0 && userID == 0 {
			// Row in google_user_ids exists, row in users doesn't: Previous registration left
			// off partway through. Continue.
		}

		if _, err := db.ExecContext(
			r.Context(),
			`
			INSERT INTO google_links (
				access_token,
				token_type,
				refresh_token,
				expires_at
			) VALUES (
				$1, $2, $3, $4, $5
			)
			`,
			token.AccessToken,
			token.TokenType,
			token.RefreshToken,
			token.Expiry,
			userinfo.Id,
		); err != nil {
			logger.Printf("unable to insert google link: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		googleUserResult, err := db.ExecContext(
			r.Context(),
			`
	 		INSERT INTO google_users (
				external_id,
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
			logger.Printf("unable to insert google user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		googleUserID, err := googleUserResult.LastInsertId()
		if err != nil {
			logger.Printf("unable to get google user id: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err := db.ExecContext(
			r.Context(),
			`
			UPDATE google_links
			SET google_user_id = $1
			WHERE id = $2
			`,
			googleUserID,
		); err != nil {
			logger.Printf("unable to insert google user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userResult, err := db.ExecContext(
			r.Context(),
			`
			INSERT INTO users (
			) VALUES (
			)
			`,
		)
		if err != nil {
			logger.Printf("unable to insert user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userID, err := userResult.LastInsertId()
		if err != nil {
			logger.Printf("unable to get user id: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err := db.ExecContext(
			r.Context(),
			`
			UPDATE google_users
			SET user_id = $1
			WHERE id = $2
			`,
			userID,
			googleUserID,
		); err != nil {
			logger.Printf("unable to insert google user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	return &googleClient{
		http: &http.Client{},
	}, nil
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
	var err error
	token.Expiry, err = time.Parse(time.RFC3339, expiry)
	if err != nil {
		return nil, fmt.Errorf("invalid datetime in Google link expires_at column `%s`: %w", expiry, err)
	}

	return &token, nil
}

func generateState(ctx context.Context, db *sql.DB) (string, error) {
	data := make([]byte, 1024)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return "", err
	}
	s := base64.StdEncoding.EncodeToString(data)
	if _, err := db.ExecContext(
		ctx,
		`
		INSERT INTO google_oauth_states (
			state
		) VALUES (
			$1
		)
		`,
		s,
	); err != nil {
		return "", err
	}
	return s, nil
}

func consumeState(ctx context.Context, db *sql.DB, state string) error {
	row := db.QueryRowContext(
		ctx,
		`
		SELECT id, created_at
		FROM google_oauth_states
		WHERE state = $1
		`,
		state,
	)
	var (
		id           int
		createdAtStr string
	)
	if err := row.Scan(&id, &createdAtStr); err != nil {
		return fmt.Errorf("cannot scan row: %w", err)
	}
	if id == 0 {
		return errors.New("state not found")
	}
	if _, err := db.ExecContext(
		ctx,
		`
		DELETE FROM google_oauth_states
		WHERE id = $1
		`,
		id,
	); err != nil {
		return fmt.Errorf("cannot delete state: %w", err)
	}

	createdAt, err := time.Parse(time.DateTime, createdAtStr)
	if err != nil {
		return fmt.Errorf("cannot parse created_at: %w", err)
	}

	if time.Since(createdAt) > 5*time.Minute {
		return errors.New("state expired")
	}

	return nil
}
