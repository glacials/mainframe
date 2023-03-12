package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"twos.dev/mainframe/coldbrewcrew/iworkout"
)

const port = 9000

//go:embed html
var html embed.FS

//go:embed static
var static embed.FS

// IworkoutParams are the fields sent to the template which renders
// html/iworkout.html.
type IworkoutParams struct {
	// Users is a map of all users we've catalogued, as a map of user ID to user.
	Users map[string]discordgo.User

	// Messages is a map of message IDs to messages.
	Messages map[string]discordgo.Message

	// Reactions is a map of message IDs to users who reacted to that message.
	Reactions map[string]map[string]struct{}
}

func overrideMIMEType(logger *log.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("overriding MIME type for %s", r.URL.Path)
		w.Header().Set("Content-Type", "text/css")
		h.ServeHTTP(w, r)
	})
}

// Start boots the web server in a goroutine and then immediately returns the
// root serve mux.
func Start(logger *log.Logger, version string) (*http.ServeMux, error) {
	logger = log.New(logger.Writer(), "[web] ", logger.Flags())
	logger.Println("Booting web")

	htmlfs, err := fs.Sub(html, "html")
	if err != nil {
		return nil, fmt.Errorf("failed to get html subdirectory: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", overrideMIMEType(logger, http.FileServer(http.FS(static))))
	mux.Handle("/", http.FileServer(http.FS(htmlfs)))
	t, err := template.ParseFS(htmlfs, "*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}
	mux.HandleFunc("/iworkout", func(w http.ResponseWriter, r *http.Request) {
		params := IworkoutParams{
			Users:     iworkout.Users(),
			Messages:  iworkout.Messages(),
			Reactions: iworkout.Reactions(),
		}
		if err := t.Lookup("iworkout.html.tmpl").Execute(w, params); err != nil {
			logger.Printf("error executing iworkout template: %s", err)
			w.WriteHeader(500)
			return
		}
	})

	logger.Printf("Booted web, listening on http://%s:%d\n", "localhost", port)
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
			logger.Printf("server stopped: %v", err)
		}
	}()

	return mux, nil
}

// AddOne is a template convenience function that returns 1 + its argument.
func AddOne(i int) int {
	return i + 1
}
