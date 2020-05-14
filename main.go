package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/glacials/mainframe/coldbrewcrew/iworkout"
)

const startMsg = "Starting mainframe..."
const port = 9000
const refreshInterval = 2 * time.Second

func main() {
	var (
		index []byte
		err   error
	)
	go func() {
		for {
			index, err = ioutil.ReadFile("html/index.html")
			if err != nil {
				log.Fatalf("Cannot read html/index.html: %s", err)
				return
			}

			time.Sleep(refreshInterval)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	})

	fmt.Printf("%s listening on %s:%d.\n", startMsg, "localhost", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
