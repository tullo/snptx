package snippet

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific Snippet is requested but does not exist.
	ErrNotFound = errors.New("Snippet not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Create inserts a new snippet record into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewSnippet, now time.Time) (*Snippet, error) {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Create")
	defer span.End()

	s := Snippet{
		ID:          uuid.New().String(),
		Title:       n.Title,
		Content:     n.Content,
		DateExpires: n.DateExpires,
		DateCreated: now,
		DateUpdated: now,
	}

	const q = `INSERT INTO snippets
	(snippet_id, title, content, date_expires, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.ExecContext(ctx, q,
		s.ID, s.Title, s.Content, s.DateExpires, s.DateCreated, s.DateUpdated,
	)
	if err != nil {
		return nil, errors.Wrap(err, "inserting snippet")
	}

	return &s, nil
}

// Retrieve gets the specified snippet from the database.
func Retrieve(ctx context.Context, db *sqlx.DB, id string) (*Snippet, error) {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var s Snippet
	const q = `SELECT * FROM snippets WHERE snippet_id = $1`
	if err := db.GetContext(ctx, &s, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting snippet %q", id)
	}

	return &s, nil
}

// Update updates a snippet record in the database.
func Update(ctx context.Context, db *sqlx.DB, id string, upd UpdateSnippet, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Update")
	defer span.End()

	s, err := Retrieve(ctx, db, id)
	if err != nil {
		return err
	}

	if upd.Title != nil {
		s.Title = *upd.Title
	}
	if upd.Content != nil {
		s.Content = *upd.Content
	}
	if upd.DateExpires != nil {
		s.DateExpires = *upd.DateExpires
	}

	s.DateUpdated = now

	const q = `UPDATE snippets SET
		"title" = $2,
		"content" = $3,
		"date_expires" = $4,
		"date_updated" = $5
		WHERE snippet_id = $1`
	_, err = db.ExecContext(ctx, q, id, s.Title, s.Content, s.DateExpires, now)
	if err != nil {
		return errors.Wrap(err, "updating snippet")
	}

	return nil
}

// Delete removes a snippet record from the database.
func Delete(ctx context.Context, db *sqlx.DB, id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Delete")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM snippets WHERE snippet_id = $1`

	if _, err := db.ExecContext(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting snippet %s", id)
	}

	return nil
}

// Latest gets the latest snippets from the database.
func Latest(ctx context.Context, db *sqlx.DB) ([]Snippet, error) {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Latest")
	defer span.End()

	snippets := []Snippet{}
	const q = `SELECT * FROM snippets
		WHERE date_expires > NOW() ORDER BY date_created DESC LIMIT 10;`
	if err := db.SelectContext(ctx, &snippets, q); err != nil {
		return nil, errors.Wrap(err, "selecting snippets")
	}

	return snippets, nil
}
