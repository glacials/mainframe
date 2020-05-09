package main

import "fmt"

type Person interface {
	FullName() string
}

type personImpl struct {
	Username  string
	FullNames []string
	Email     string
}

func (p personImpl) FullName() string {
	return p.FullNames[0]
}

func main() {
	person := personImpl{
		FullNames: []string{"Ben Carlsson", "Benjamin Carlsson", "Benjamin Nicholas Carlsson"},
	}

	register(person)
}

func register(p Person) {
	fmt.Printf("Registered %s", p.FullName())
}
