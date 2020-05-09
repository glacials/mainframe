// person holds logic and information about a person registered to the Outreach database.
package person

import (
	"fmt"

	"github.com/google/uuid"
)

type Person interface {
	ID() uuid.UUID
	FullName() string
}

// Person represents a person in the database.
type personImpl struct {
	id    uuid.UUID
	name  string
	email string
}

func (p personImpl) ID() uuid.UUID {
	return p.id
}

func (p personImpl) FullName() string {
	return p.name
}

// Return structs, accept interfaces

// New returns a new Person.
func New(name, email string) personImpl {
	return personImpl{
		id:    uuid.New(),
		name:  name,
		email: email,
	}
}

// String returns a Person as a string.
func (p personImpl) String() string {
	return fmt.Sprintf("Person[%s]", p.name)
}
