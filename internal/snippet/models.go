package snippet

import (
	"time"
)

// Info represents a textual extract of something
type Info struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	DateExpires time.Time `json:"date_expires"`
	DateCreated time.Time `json:"date_created"`
	DateUpdated time.Time `json:"date_updated"`
}

// NewSnippet contains information needed to create a new Snippet.
type NewSnippet struct {
	Title       string    `json:"title" validate:"required"`
	Content     string    `json:"content" validate:"required"`
	DateExpires time.Time `json:"date_expires" validate:"required"`
}

// UpdateSnippet defines what information may be provided to modify an existing
// Snippet. All fields are optional so clients can send just the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicitly blank.
type UpdateSnippet struct {
	Title       *string    `json:"title"`
	Content     *string    `json:"content"`
	DateExpires *time.Time `json:"date_expires"`
}
