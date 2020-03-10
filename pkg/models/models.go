package models

import (
	"errors"
	"time"
)

// ErrNoRecord is used when a specific snippet is requested but does not exist.
var ErrNoRecord = errors.New("models: no matching record found")

// Snippet is an item we manage
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}
