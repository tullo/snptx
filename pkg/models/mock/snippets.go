package mock

import (
	"fmt"
	"time"

	"github.com/tullo/snptx/pkg/models"
)

var mockSnippet = &models.Snippet{
	ID:      1,
	Title:   "An old silent pond",
	Content: "An old silent pond...",
	Created: time.Now(),
	Expires: time.Now(),
}

// SnippetModel ...
type SnippetModel struct{}

// Insert ...
func (m *SnippetModel) Insert(title, content, expires string) (int, error) {
	return 2, nil
}

// Get ..
func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	switch id {
	case 1:
		return mockSnippet, nil
	case 66:
		return nil, fmt.Errorf("Internal Server Error")
	default:
		return nil, models.ErrNoRecord
	}
}

// Latest ...
func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	return []*models.Snippet{mockSnippet}, nil
}
