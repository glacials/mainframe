// Package person manages people.
package person

import "fmt"

type person struct {
	firstName string
	lastName  string
	admin     bool
}

// New creates a new person and returns it.
func New(firstName, lastName string) person {
	return person{
		firstName: firstName,
		lastName:  lastName,
	}
}

// FullName returns the full name of the person.
func (p person) FullName() string {
	return fmt.Sprintf("%s %s", p.firstName, p.lastName)
}

// destroy destroys the person.
func (p person) destroy() {
	// ...
}
