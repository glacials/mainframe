package main

import "fmt"

type Person struct {
	Username string
	FullName string
	Email    string
}

func main() {
	person := Person{
		Username: "glacials",
		FullName: "Ben Carlsson",
		Email:    "ben@twos.dev",
	}

	fmt.Printf("%v\n", person)
}
