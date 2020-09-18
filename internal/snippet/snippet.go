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

// Snippet manages the set of API's for snippet access. It wraps a sql.DB
// connection pool.
type Snippet struct {
	db *sqlx.DB
}

// New constructs a Snippet for api access.
func New(db *sqlx.DB) Snippet {
	return Snippet{db: db}
}

// Create inserts a new snippet record into the database.
func (s Snippet) Create(ctx context.Context, n NewSnippet, now time.Time) (*Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Create")
	defer span.End()

	spt := Info{
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
	_, err := s.db.ExecContext(ctx, q,
		spt.ID, spt.Title, spt.Content, spt.DateExpires, spt.DateCreated, spt.DateUpdated,
	)
	if err != nil {
		return nil, errors.Wrap(err, "inserting snippet")
	}

	return &spt, nil
}

// Retrieve gets the specified snippet from the database.
func (s Snippet) Retrieve(ctx context.Context, id string) (*Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var spt Info
	const q = `SELECT * FROM snippets WHERE snippet_id = $1`
	if err := s.db.GetContext(ctx, &spt, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting snippet %q", id)
	}

	return &spt, nil
}

// Update updates a snippet record in the database.
func (s Snippet) Update(ctx context.Context, id string, upd UpdateSnippet, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Update")
	defer span.End()

	spt, err := s.Retrieve(ctx, id)
	if err != nil {
		return err
	}

	if upd.Title != nil {
		spt.Title = *upd.Title
	}
	if upd.Content != nil {
		spt.Content = *upd.Content
	}
	if upd.DateExpires != nil {
		spt.DateExpires = *upd.DateExpires
	}

	spt.DateUpdated = now

	const q = `UPDATE snippets SET
		"title" = $2,
		"content" = $3,
		"date_expires" = $4,
		"date_updated" = $5
		WHERE snippet_id = $1`
	_, err = s.db.ExecContext(ctx, q, id, spt.Title, spt.Content, spt.DateExpires, now)
	if err != nil {
		return errors.Wrap(err, "updating snippet")
	}

	return nil
}

// Delete removes a snippet record from the database.
func (s Snippet) Delete(ctx context.Context, id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Delete")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM snippets WHERE snippet_id = $1`

	if _, err := s.db.ExecContext(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting snippet %s", id)
	}

	return nil
}

// Latest gets the latest snippets from the database.
func (s Snippet) Latest(ctx context.Context) ([]Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Latest")
	defer span.End()

	snippets := []Info{}
	const q = `SELECT * FROM snippets
		WHERE date_expires > NOW() ORDER BY date_created DESC LIMIT 10;`
	if err := s.db.SelectContext(ctx, &snippets, q); err != nil {
		return nil, errors.Wrap(err, "selecting snippets")
	}

	return snippets, nil
}
