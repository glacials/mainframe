// Package db is a simple file-based key-value store. It implictly trusts its consumers not to be malicious to the filesystem, such as using ../ to navigate out of the database directory.
package db

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

const prefix = "db/store"

// GetStr retrieves a string value from the store and returns it.
func GetStr(key string) (string, error) {
	val, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", prefix, key))
	if err != nil {
		return "", fmt.Errorf("Can't get value %s from db: %s", key, err)
	}

	return string(val), nil
}

// SetStr places a key in the store with string value val.
func SetStr(key, val string) error {
	if err := ioutil.WriteFile(fmt.Sprintf("%s/%s", prefix, key), []byte(val), os.ModePerm); err != nil {
		return fmt.Errorf("Can't set value %s=%s in db: %s", key, val, err)
	}

	return nil
}

// GetTime retrieves a time.Time value from the store and returns it.
func GetTime(key string) (time.Time, error) {
	val, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", prefix, key))
	if err != nil {
		return time.Time{}, fmt.Errorf("Can't get time.Time value %s from db: %s", key, err)
	}

	var t time.Time

	t.UnmarshalText([]byte(val))

	return t, nil
}
func SetTime(key string, val time.Time) error {
	t, err := val.MarshalText()
	if err != nil {
		return fmt.Errorf("Can't marshal time.Time %v into text: %s", val, err)
	}

	if err := ioutil.WriteFile(fmt.Sprintf("%s/%s", prefix, key), t, os.ModePerm); err != nil {
		return fmt.Errorf("Can't set time.Time value %s=%v in db: %s", key, val, err)
	}

	return nil
}

// SetStr places a key in the store with time.Time value val.
