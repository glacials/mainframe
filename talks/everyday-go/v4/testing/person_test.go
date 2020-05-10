package person

// person_test.go
import "testing"

func TestFullName(t *testing.T) {
	expected := "Ben Carlsson"
	actual := person{firstName: "Ben", lastName: "Carlsson"}.FullName()
	if expected != actual {
		t.Fail()
	}
}
