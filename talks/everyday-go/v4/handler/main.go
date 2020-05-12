// main.go
package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(401)
	w.Write([]byte("Hello, world!\n"))
}
