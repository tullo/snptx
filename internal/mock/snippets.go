package mock

import (
	"fmt"
	"time"

	"github.com/tullo/snptx/internal/snippet"
	"github.com/tullo/snptx/pkg/models"
)

var mockSnippet = &snippet.Snippet{
	ID:          "1",
	Title:       "An old silent pond",
	Content:     "An old silent pond...",
	DateCreated: time.Now(),
	DateExpires: time.Now(),
}

// SnippetModel ...
type SnippetModel struct{}

// Insert ...
func (m *SnippetModel) Insert(title, content, expires string) (string, error) {
	return "2", nil
}

// Get ..
func (m *SnippetModel) Get(id string) (*snippet.Snippet, error) {
	switch id {
	case "1":
		return mockSnippet, nil
	case "66":
		return nil, fmt.Errorf("Internal Server Error")
	default:
		return nil, models.ErrNoRecord
	}
}

// Latest ...
func (m *SnippetModel) Latest() ([]snippet.Snippet, error) {
	return []snippet.Snippet{*mockSnippet}, nil
}

// Update ..
func (m *SnippetModel) Update(id string, up snippet.UpdateSnippet) error {
	switch id {
	case "1":
		return nil
	case "66":
		return fmt.Errorf("Internal Server Error")
	default:
		return models.ErrNoRecord
	}
}
