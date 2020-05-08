package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const port = 9000
const refreshInterval = 2 * time.Second

func main() {
	fmt.Println("Starting mainframe...")

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

	fmt.Println("Mainframe listening on port 8080...")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
