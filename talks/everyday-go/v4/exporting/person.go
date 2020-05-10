// person.go
package person

import "fmt"

type person struct {
	firstName string
	lastName  string
	admin     bool
}

func New(firstName, lastName string) person {
	return person{
		firstName: firstName,
		lastName:  lastName,
	}
}

func (p person) FullName() string {
	return fmt.Sprintf("%s %s", p.firstName, p.lastName)
}

func (p person) destroy() { ... }
