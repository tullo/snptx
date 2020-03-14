package snippet

import (
	"time"
)

// Snippet represents a textual extract of something
type Snippet struct {
	ID          string    `db:"snippet_id" json:"id"`
	Title       string    `db:"title" json:"title"`
	Content     string    `db:"content" json:"content"`
	DateExpires time.Time `db:"date_expires" json:"date_expires"`
	DateCreated time.Time `db:"date_created" json:"date_created"`
	DateUpdated time.Time `db:"date_updated" json:"date_updated"`
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
