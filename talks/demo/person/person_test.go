package person

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestPersonString(t *testing.T) {
	p := Person{
		ID:    uuid.New(),
		Name:  "Testy McTesterson",
		Email: "test@test.com",
	}

	expected := fmt.Sprintf("Person[%s]", p.Name)
	actual := p.String()

	if actual != expected {
		t.Fatalf("Expected person name to be %s, was %s", expected, actual)
	}
}
