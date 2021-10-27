package web

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/glacials/mainframe/coldbrewcrew/iworkout"
)

const port = 80

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

// Start boots the web server and listens for connections.
func Start(logger *log.Logger) error {
	logger = log.New(logger.Writer(), "[web] ", logger.Flags())
	logger.Println("Booting web")

	templates := template.New("html")
	templates = templates.Funcs(map[string]interface{}{
		"AddOne": AddOne,
	})

	templates, err := templates.ParseFS(html, "html/*")
	if err != nil {
		return fmt.Errorf("cannot parse template: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := templates.ExecuteTemplate(w, "index.html", struct{}{}); err != nil {
			logger.Printf("Error executing template: %s", err)
			w.WriteHeader(500)
		}
	})

	http.HandleFunc("/iworkout", func(w http.ResponseWriter, r *http.Request) {
		params := IworkoutParams{
			Users:     iworkout.Users(),
			Reactions: iworkout.Reactions(),
			Messages:  iworkout.Messages(),
		}
		if err := templates.ExecuteTemplate(w, "iworkout.html", params); err != nil {
			logger.Printf("Error executing template: %s", err)
			w.WriteHeader(500)
		}
	})

	logger.Printf("Booted web, listening on %s:%d\n", "localhost", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		return fmt.Errorf("server stopped: %v", err)
	}

	return nil
}

// AddOne is a template convenience function that returns 1 + its argument.
func AddOne(i int) int {
	return i + 1
}
