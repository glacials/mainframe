package user

import "github.com/google/uuid"

type user struct {
	id   uuid.UUID
	name string
}

func New(name string) user {
	return user{
		id:   uuid.New(),
		name: name,
	}
}
