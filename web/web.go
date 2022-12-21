package web

import (
	"embed"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	"github.com/bwmarrin/discordgo"
	"github.com/markbates/pkger"
	"twos.dev/mainframe/coldbrewcrew/iworkout"
)

const port = 9000

//go:embed html
var html embed.FS

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
func Start(logger *log.Logger) (*http.ServeMux, error) {
	logger = log.New(logger.Writer(), "[web] ", logger.Flags())
	logger.Println("Booting web")

	dir := http.FileServer(pkger.Dir("/web/html/static"))
	dir = overrideMIMEType(logger, dir)

	indexFile, err := pkger.Open("/web/html/index.html")
	if err != nil {
		return nil, fmt.Errorf("can't open index.html: %v", err)
	}

	indexBytes, err := ioutil.ReadAll(indexFile)
	if err != nil {
		return nil, fmt.Errorf("can't read index.html: %v", err)
	}

	indexStr := string(indexBytes)
	indexTemplate := template.Must(template.New("").Parse(indexStr))

	iworkoutFile, err := pkger.Open("/web/html/iworkout.html")
	if err != nil {
		return nil, fmt.Errorf("can't open iworkout.html: %v", err)
	}

	iworkoutBytes, err := ioutil.ReadAll(iworkoutFile)
	if err != nil {
		return nil, fmt.Errorf("can't read iworkout.html: %v", err)
	}

	iworkoutStr := string(iworkoutBytes)
	iworkoutTemplate := template.Must(template.New("").Parse(iworkoutStr))

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", dir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := indexTemplate.Execute(w, nil); err != nil {
			logger.Printf("error executing index template: %v", err)
		}
	})
	mux.HandleFunc("/iworkout", func(w http.ResponseWriter, r *http.Request) {
		params := IworkoutParams{
			Users:     iworkout.Users(),
			Messages:  iworkout.Messages(),
			Reactions: iworkout.Reactions(),
		}
		if err := iworkoutTemplate.Execute(w, params); err != nil {
			logger.Printf("error executing iworkout template: %s", err)
			w.WriteHeader(500)
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
