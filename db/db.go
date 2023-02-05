// Package db is a utility library to assist in interacting with SQLite. It is
// not an abstraction layer.
package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/pkger"
	_ "modernc.org/sqlite"
)

// New creates a database connection to the SQLite database at the given path,
// migrates the database if necessary, and returns the connection.
//
// It is the caller's responsibility to call db.Close().
func New(logger *log.Logger, name string) (*sql.DB, error) {
	path := fmt.Sprintf("%s.db", name)

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %v", err)
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{DatabaseName: name})
	if err != nil {
		return nil, fmt.Errorf(
			"can't create migration driver for %s: %v",
			path,
			err,
		)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"pkger:///db/migrations",
		name,
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("can't set up migrations: %v", err)
	}

	err = m.Up()
	if err == migrate.ErrNoChange {
		return db, nil
	} else if err != nil {
		return nil, fmt.Errorf("can't migrate database: %v", err)
	}

	logger.Printf("database migrated")
	return db, nil
}
