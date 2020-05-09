package main

import "fmt"

type Username string
type FullName string

func main() {
	var (
		username Username = "glacials"
		fullname FullName = "Ben Carlsson"
	)

	register(username, fullname)
	//register(fullname, username)
	//register("glacials", "Ben Carlsson")
}

func register(u Username, f FullName) {
	fmt.Printf("Registered username %s as %s\n", u, f)
}
