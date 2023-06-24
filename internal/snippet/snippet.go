package snippet

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific Snippet is requested but does not exist.
	ErrNotFound = errors.New("Snippet not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Store manages the set of API's for snippet access. It wraps a pgxpool.Pool
// connection pool.
type Store struct {
	db *pgxpool.Pool
}

// NewStore constructs a Store for api access.
func NewStore(db *pgxpool.Pool) Store {
	return Store{db: db}
}

// Create inserts a new snippet record into the database.
func (s Store) Create(ctx context.Context, n NewSnippet, now time.Time) (*Info, error) {
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

	const q = `
	INSERT INTO snippets
	  (snippet_id, title, content, date_expires, date_created, date_updated)
	VALUES
	  ($1, $2, $3, $4, $5, $6)`
	if _, err := s.db.Exec(ctx, q, spt.ID, spt.Title, spt.Content, spt.DateExpires, spt.DateCreated, spt.DateUpdated); err != nil {
		return nil, errors.Wrap(err, "inserting snippet")
	}

	return &spt, nil
}

// Retrieve gets the specified snippet from the database.
func (s Store) Retrieve(ctx context.Context, id string) (*Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var spt Info
	const q = `SELECT * FROM snippets WHERE snippet_id = $1`
	if err := pgxscan.Get(ctx, s.db, &spt, q, id); err != nil {
		if pgxscan.NotFound(err) {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting snippet %q", id)
	}

	return &spt, nil
}

// Update updates a snippet record in the database.
func (s Store) Update(ctx context.Context, id string, upd UpdateSnippet, now time.Time) error {
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

	const q = `
	UPDATE snippets SET
	  "title" = $2,
	  "content" = $3,
	  "date_expires" = $4,
	  "date_updated" = $5
	WHERE snippet_id = $1`
	if _, err = s.db.Exec(ctx, q, id, spt.Title, spt.Content, spt.DateExpires, now); err != nil {
		return errors.Wrap(err, "updating snippet")
	}

	return nil
}

// Delete removes a snippet record from the database.
func (s Store) Delete(ctx context.Context, id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Delete")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM snippets WHERE snippet_id = $1`
	if _, err := s.db.Exec(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting snippet %s", id)
	}

	return nil
}

// Latest gets the latest snippets from the database.
func (s Store) Latest(ctx context.Context) ([]Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Latest")
	defer span.End()

	snippets := []Info{}
	const q = `
	SELECT * FROM snippets
	WHERE date_expires > NOW()
	ORDER BY date_created DESC
	LIMIT 10;`
	if err := pgxscan.Select(ctx, s.db, &snippets, q); err != nil {
		return nil, errors.Wrap(err, "selecting snippets")
	}

	return snippets, nil
}
