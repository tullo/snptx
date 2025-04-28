package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/tullo/snptx/internal/models"
	"github.com/tullo/snptx/internal/snippet"
)

var mockSnippet = &models.Snippet{
	ID:          "1",
	Title:       "An old silent pond",
	Content:     "An old silent pond...",
	DateCreated: time.Now(),
	DateExpires: time.Now(),
}

// SnippetStore manages the set of API's for snippet access
type SnippetStore struct{}

// NewSnippetStore constructs a SnippetStore for api access.
func NewSnippetStore() SnippetStore {
	var s SnippetStore
	return s
}

// Create inserts a new snippet record into the database.
func (s SnippetStore) Create(context.Context, models.NewSnippet, time.Time) (*models.Snippet, error) {
	var spt models.Snippet
	spt.ID = "2"
	return &spt, nil
}

// Retrieve gets the specified snippet from the database.
func (s SnippetStore) Retrieve(ctx context.Context, id string) (*models.Snippet, error) {
	switch id {
	case "1":
		return mockSnippet, nil
	case "66":
		return nil, fmt.Errorf("internal server error")
	default:
		return nil, models.ErrNoRecord
	}
}

// Latest gets the latest snippets from the database.
func (s SnippetStore) Latest(context.Context) ([]models.Snippet, error) {
	return []models.Snippet{*mockSnippet}, nil
}

// Update updates a snippet record in the database.
func (s SnippetStore) Update(ctx context.Context, id string, us models.UpdateSnippet, t time.Time) error {
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
func (s SnippetStore) Delete(context.Context, string) error {
	return nil
}
