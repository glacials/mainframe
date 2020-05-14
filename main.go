package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/glacials/mainframe/coldbrewcrew/iworkout"
	_ "github.com/glacials/mainframe/coldbrewcrew/iworkout"
)

const startMsg = "Starting mainframe..."
const port = 9000
const refreshInterval = 1 * time.Second

var iworkoutTemplate *template.Template

func main() {
	var (
		index, iworkoutFile []byte
		err                 error
	)
	go func() {
		for {
			index, err = ioutil.ReadFile("html/index.html")
			if err != nil {
				log.Fatalf("Cannot read html/index.html: %s", err)
				return
			}

			iworkoutFile, err = ioutil.ReadFile("html/iworkout.html")
			if err != nil {
				log.Fatalf("Cannot read html/index.html: %s", err)
				return
			}

			iworkoutTemplate = template.New("iworkout")
			iworkoutTemplate = iworkoutTemplate.Funcs(map[string]interface{}{
				"AddOne": AddOne,
			})

			iworkoutTemplate, err = iworkoutTemplate.Parse(string(iworkoutFile))
			if err != nil {
				log.Printf("Cannot parse template html/iworkout.html: %s", err)
				return
			}

			time.Sleep(refreshInterval)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	})

	type IworkoutParams struct {
		// Users is a map of all users we've catalogued, as a map of user ID to user.
		Users map[string]discordgo.User

		// Messages is a map of message IDs to messages.
		Messages map[string]discordgo.Message

		// Reactions is a map of message IDs to users who reacted to that message.
		Reactions map[string]map[string]struct{}
	}

	http.HandleFunc("/iworkout", func(w http.ResponseWriter, r *http.Request) {
		params := IworkoutParams{
			Users:     iworkout.Users(),
			Reactions: iworkout.Reactions(),
			Messages:  iworkout.Messages(),
		}
		if err := iworkoutTemplate.Execute(w, params); err != nil {
			log.Printf("Error executing template: %s", err)
		}
	})

	fmt.Printf("%s listening on %s:%d.\n", startMsg, "localhost", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func AddOne(i int) int {
	return i + 1
}
