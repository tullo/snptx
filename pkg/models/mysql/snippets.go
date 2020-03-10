package mysql

import (
	"database/sql"
	"errors"

	"github.com/tullo/snptx/pkg/models"
)

// SnippetModel wraps a sql.DB connection pool
type SnippetModel struct {
	DB *sql.DB
}

// Insert knows how to insert a new snippet into the database
func (m *SnippetModel) Insert(title, content, expires string) (int, error) {

	stmt := `INSERT INTO snippets (title, content, created, expires)
	    VALUES(
			?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY
		)
	)`
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Get knows how to retreive a specific snippet based on its id
func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	s := &models.Snippet{}

	stmt := `SELECT id, title, content, created, expires
		FROM snippets
		WHERE expires > UTC_TIMESTAMP() AND id = ?`

	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		// unwrapping errors
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		}
		return nil, err
	}

	return s, nil
}

// Latest knows how to retreive the 10 most recently created snippets
func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	stmt := `SELECT id, title, content, created, expires
		FROM snippets
		WHERE expires > UTC_TIMESTAMP() ORDER BY created DESC LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snippets := []*models.Snippet{}

	for rows.Next() {
		s := &models.Snippet{}
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
