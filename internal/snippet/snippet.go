package snippet

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tullo/snptx/internal/db"
	"github.com/tullo/snptx/internal/platform/database"
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
	db *database.DB
	q  *db.Queries
}

// NewStore constructs a Store for api access.
func NewStore(d *database.DB) Store {
	return Store{
		db: d,
		q:  db.New(d),
	}
}

// Create inserts a new snippet record into the database.
func (s Store) Create(ctx context.Context, n NewSnippet, now time.Time) (*Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Create")
	defer span.End()

	sn, err := s.q.CreateSnippet(ctx, db.GetCreateSnippetParams(
		uuid.New().String(),
		n.Title,
		n.Content,
		n.DateExpires,
		now,
		now,
	))
	if err != nil {
		return nil, errors.Wrap(err, "inserting snippet")
	}

	spt := Info{
		ID:          sn.SnippetID,
		Title:       sn.Title.String,
		Content:     sn.Content.String,
		DateExpires: sn.DateExpires.Time,
		DateCreated: sn.DateCreated.Time,
		DateUpdated: sn.DateUpdated.Time,
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

	snip, err := s.q.GetSnippet(ctx, id)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting snippet %q", id)
	}

	return &Info{
		ID:          snip.SnippetID,
		Title:       snip.Title.String,
		Content:     snip.Content.String,
		DateExpires: snip.DateExpires.Time,
		DateCreated: snip.DateCreated.Time,
		DateUpdated: snip.DateUpdated.Time,
	}, nil
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

	err = s.q.UpdateSnippet(ctx, db.GetUpdateSnippetParams(
		id,
		spt.Title,
		spt.Content,
		spt.DateExpires,
		spt.DateUpdated,
	))
	if err != nil {
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

	err := s.q.DeleteSnippet(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "deleting snippet %s", id)
	}

	return nil
}

// Latest gets the latest snippets from the database.
func (s Store) Latest(ctx context.Context) ([]Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Latest")
	defer span.End()

	ss, err := s.q.ListLatestSnippets(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "selecting snippets")
	}

	is := make([]Info, len(ss))
	for i, v := range ss {
		is[i] = Info{
			ID:          v.SnippetID,
			Title:       v.Title.String,
			Content:     v.Content.String,
			DateExpires: v.DateExpires.Time,
			DateCreated: v.DateCreated.Time,
			DateUpdated: v.DateUpdated.Time,
		}
	}

	return is, nil
}
