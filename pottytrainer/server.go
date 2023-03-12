package pottytrainer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"twos.dev/mainframe/ddb"
)

// Run attaches all the right handlers to the given serve mux.
func Run(logger *log.Logger, mux *http.ServeMux) error {
	logger = log.New(logger.Writer(), "[potty] ", logger.Flags())

	ddb, err := ddb.New(&ddb.Config{
		Region:          "us-east-1",
		TableNamePrefix: "mainframe-potty",
	})
	if err != nil {
		return fmt.Errorf("failed to create ddb client: %w", err)
	}

	apiMux := http.NewServeMux()
	apiMux.Handle("/eat", wrapHandler(eatHandler, logger, ddb))
	apiMux.Handle("/poop", wrapHandler(poopHandler, logger, ddb))

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiMux))
	return nil
}

type customHandler func(http.ResponseWriter, *http.Request, *log.Logger, *ddb.Client, *ddb.User)

func wrapHandler(h customHandler, logger *log.Logger, db *ddb.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var token string
		token = r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		if token == "" {
			http.Error(
				w,
				fmt.Sprintf(
					"missing token, please pass it like `%s` or `%s`",
					"?token=changeme",
					"Authorization: Bearer changeme",
				),
				http.StatusUnauthorized,
			)
			return
		}

		user, err := db.UserFromToken(ctx, token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		h(w, r, logger, db, user)
	})
}

type eatRequest struct {
	AteAt time.Time `json:"ate_at"`
}

// eatHandler returns a handler that handles the /eat route in the /api/v1
// namespace. It implements customHandler.
func eatHandler(w http.ResponseWriter, r *http.Request, logger *log.Logger, ddb *ddb.Client, user *ddb.User) {
	//ctx := r.Context()

	//var e eatRequest
	if _, err := w.Write([]byte("yum")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type poopRequest struct {
	PoopedAt time.Time `json:"pooped_at"`
	Bad      bool      `json:"quality"`
}

// poopHandler returns a handler that handles the /poop route in the /api/v1
// namespace. It implements customHandler.
func poopHandler(w http.ResponseWriter, r *http.Request, logger *log.Logger, db *ddb.Client, user *ddb.User) {
	ctx := r.Context()

	var p poopRequest
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := LogPoop(ctx, db, user.ID, p.PoopedAt, p.Bad); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
