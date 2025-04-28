package models

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

type SnippetModelInterface interface {
	Create(context.Context, NewSnippet, time.Time) (*Snippet, error)
	Delete(context.Context, string) error
	Latest(context.Context) ([]Snippet, error)
	Update(context.Context, string, UpdateSnippet, time.Time) error
	Retrieve(context.Context, string) (*Snippet, error)
}

// Store manages the set of API's for snippet access. It wraps a pgxpool.Pool
// connection pool.
type SnippetStore struct {
	db *database.DB
	q  *db.Queries
}

// NewStore constructs a Store for api access.
func NewSnippetStore(d *database.DB) SnippetStore {
	return SnippetStore{
		db: d,
		q:  db.New(d),
	}
}

// Create inserts a new snippet record into the database.
func (s SnippetStore) Create(ctx context.Context, n NewSnippet, now time.Time) (*Snippet, error) {
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

	spt := Snippet{
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
func (s SnippetStore) Retrieve(ctx context.Context, id string) (*Snippet, error) {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	snip, err := s.q.GetSnippet(ctx, id)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, ErrNoRecord
		}

		return nil, errors.Wrapf(err, "selecting snippet %q", id)
	}

	return &Snippet{
		ID:          snip.SnippetID,
		Title:       snip.Title.String,
		Content:     snip.Content.String,
		DateExpires: snip.DateExpires.Time.In(copenhagen),
		DateCreated: snip.DateCreated.Time.In(copenhagen),
		DateUpdated: snip.DateUpdated.Time.In(copenhagen),
	}, nil
}

// Update updates a snippet record in the database.
func (s SnippetStore) Update(ctx context.Context, id string, upd UpdateSnippet, up time.Time) error {
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

	spt.DateUpdated = up

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
func (s SnippetStore) Delete(ctx context.Context, id string) error {
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
func (s SnippetStore) Latest(ctx context.Context) ([]Snippet, error) {
	ctx, span := trace.StartSpan(ctx, "internal.snippet.Latest")
	defer span.End()

	ss, err := s.q.ListLatestSnippets(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "selecting snippets")
	}

	is := make([]Snippet, len(ss))
	for i, v := range ss {
		is[i] = Snippet{
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
