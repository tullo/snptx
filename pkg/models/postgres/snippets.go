package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tullo/snptx/internal/snippet"
)

// SnippetModel wraps a sql.DB connection pool
type SnippetModel struct {
	DB *sqlx.DB
}

// Insert knows how to insert a new snippet into the database
func (m *SnippetModel) Insert(title, content, expires string) (string, error) {

	now := time.Now()
	var exp time.Time
	switch expires {
	case "365":
		exp = now.AddDate(1, 0, 0)
	case "7":
		exp = now.AddDate(0, 0, 7)
	case "1":
		exp = now.AddDate(0, 0, 7)
	}
	ns := snippet.NewSnippet{
		Title:       title,
		Content:     content,
		DateExpires: exp,
	}

	sCreated, err := snippet.Create(context.Background(), m.DB, ns, now)
	if err != nil {
		return "", err
	}

	return sCreated.ID, nil
}

// Get knows how to retreive a specific snippet based on its id
func (m *SnippetModel) Get(id string) (*snippet.Snippet, error) {

	s, err := snippet.Retrieve(context.Background(), m.DB, id)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Latest knows how to retreive the 10 most recently created snippets
func (m *SnippetModel) Latest() ([]snippet.Snippet, error) {

	snippets, err := snippet.Latest(context.Background(), m.DB)
	if err != nil {
		return nil, err
	}

	return snippets, nil
}

// Update knows how to update a snippet
func (m *SnippetModel) Update(id string, up snippet.UpdateSnippet) error {

	now := time.Now()
	err := snippet.Update(context.Background(), m.DB, id, up, now)
	if err != nil {
		return err
	}

	return nil
}
