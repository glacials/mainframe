package main

import (
	"fmt"

	"github.com/glacials/mainframe/talks/everyday-go/v4/interface/user"
)

// main.go
func main() {
	u := user.New("Ben", "Carlsson")
	fmt.Printf("Registered %s", u.FullName())
}
