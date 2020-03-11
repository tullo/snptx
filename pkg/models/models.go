package models

import (
	"errors"
	"time"
)

var (
	// ErrNoRecord is used when a specific snippet is requested but does not exist.
	ErrNoRecord = errors.New("models: no matching record found")
	// ErrInvalidCredentials occurs when a user  tries to login with
	// an incorrect email address or password.
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	// ErrDuplicateEmail occurs when a user tries to signup with an
	// email address that's already in use
	ErrDuplicateEmail = errors.New("models: duplicate email")
)

// Snippet is an item we manage
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// User represents someone with access to our system.
type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
	Active         bool
}
