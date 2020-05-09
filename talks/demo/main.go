// Introduction to everyday Go
// Ben Carlsson
// twos.dev

// Go is simple.

package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/glacials/mainframe/talk/demo/person"
	"github.com/google/uuid"
)

// map of string to string
// uuid -> full name
var users map[uuid.UUID]person.Person = map[uuid.UUID]person.Person{}

func main() {
	go register(
		person.New(
			"Ben Carlsson",
			"qhiiyr@gmail.com",
		),
	)

	fmt.Println("Success!")
	time.Sleep(1 * time.Second)
	fmt.Printf("%v\n", users)

}

func register(p person.Person) error {
	if rand.Intn(100) < 10 {
		return errors.New("Cannot connect to db")
	}

	users[p.ID()] = p
	return nil
}
