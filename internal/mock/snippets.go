package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/tullo/snptx/internal/snippet"
)

var mockSnippet = &snippet.Info{
	ID:          "1",
	Title:       "An old silent pond",
	Content:     "An old silent pond...",
	DateCreated: time.Now(),
	DateExpires: time.Now(),
}

// Snippet manages the set of API's for snippet access
type Snippet struct{}

// NewSnippet constructs a Snippet for api access.
func NewSnippet() Snippet {
	var s Snippet
	return s
}

// Create inserts a new snippet record into the database.
func (s Snippet) Create(context.Context, snippet.NewSnippet, time.Time) (*snippet.Info, error) {
	var spt snippet.Info
	spt.ID = "2"
	return &spt, nil
}

// Retrieve gets the specified snippet from the database.
func (s Snippet) Retrieve(ctx context.Context, id string) (*snippet.Info, error) {
	switch id {
	case "1":
		return mockSnippet, nil
	case "66":
		return nil, fmt.Errorf("internal server error")
	default:
		return nil, snippet.ErrNotFound
	}
}

// Latest gets the latest snippets from the database.
func (s Snippet) Latest(context.Context) ([]snippet.Info, error) {
	return []snippet.Info{*mockSnippet}, nil
}

// Update updates a snippet record in the database.
func (s Snippet) Update(ctx context.Context, id string, us snippet.UpdateSnippet, t time.Time) error {
	switch id {
	case "1":
		return nil
	case "66":
		return fmt.Errorf("internal server error")
	default:
		return snippet.ErrNotFound
	}
}

// Delete removes a snippet record from the database.
func (s Snippet) Delete(context.Context, string) error {
	return nil
}
