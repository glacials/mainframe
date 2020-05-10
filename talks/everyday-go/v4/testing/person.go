package person

import "fmt"

// person.go
type person struct {
	firstName string
	lastName  string
}

func (p person) FullName() string {
	return fmt.Sprintf("%s %s", p.firstName, p.lastName)
}
