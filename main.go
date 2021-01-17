package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/glacials/mainframe/coldbrewcrew/iworkout"
	_ "github.com/glacials/mainframe/coldbrewcrew/iworkout"
	"github.com/glacials/mainframe/speedtest"
)

const startMsg = "Starting mainframe..."
const port = 9000
const refreshInterval = 1 * time.Second

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

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	templates := template.New("html")
	templates = templates.Funcs(map[string]interface{}{
		"AddOne": AddOne,
	})

	templates, err := templates.ParseFS(html, "html/*")
	if err != nil {
		log.Printf("Cannot parse template: %s", err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := templates.ExecuteTemplate(w, "index.html", struct{}{}); err != nil {
			log.Printf("Error executing template: %s", err)
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
			log.Printf("Error executing template: %s", err)
			w.WriteHeader(500)
		}
	})

	go func() {
		if err := speedtest.Monitor(logger); err != nil {
			log.Printf("speedtest failed: %v", err)
		}
	}()

	log.Printf("%s listening on %s:%d.\n", startMsg, "localhost", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// AddOne is a template convenience function that returns 1 + its argument.
func AddOne(i int) int {
	return i + 1
}
