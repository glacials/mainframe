package user

import "fmt"

// user/user.go
type User interface {
	FullName() string
}

type user struct{ first, last string }

func (u user) FullName() string {
	return fmt.Sprintf("%s %s", u.first, u.last)
}

func New(first, last string) user {
	return user{first, last}
}

func Merge(a, b User) user {
	return user{} // shh
}
