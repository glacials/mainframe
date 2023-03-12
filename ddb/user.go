package ddb

import (
	"context"
	"time"
)

const (
	// usersTableBaseName is the base name of the users table. The base name is
	// prepended with the table prefix to get the full table name.
	usersTableBaseName = "users"
)

// User is a user of Potty Trainer.
type User struct {
	// ID is the unique identifier for this user. It is an opaque string.
	ID string `dynamo:"id,hash"`
	// Email is the user's email address.
	Email string `dynamo:"email"`
	// Name is the user's name. The user can change this.
	Name string `dynamo:"name"`
	// CreatedAt is the time this user record was created.
	CreatedAt time.Time `dynamo:"created_at"`
	// UpdatedAt is the time this user record was last updated.
	UpdatedAt time.Time `dynamo:"updated_at"`
}

func (db *Client) UserFromToken(ctx context.Context, token string) (*User, error) {
	var apiToken APIToken
	if err := db.tokens.Get("user_id", token).OneWithContext(ctx, &apiToken); err != nil {
		return nil, err
	}
	var user User
	if err := db.users.Get("id", apiToken.UserID).OneWithContext(ctx, &user); err != nil {
		return nil, err
	}
	return &user, nil
}
