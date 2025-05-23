// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: snippets.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createSnippet = `-- name: CreateSnippet :one
INSERT INTO snippets
  (snippet_id, title, content, date_expires, date_created, date_updated)
  VALUES
    ($1, $2, $3, $4, $5, $6)
  RETURNING snippet_id, title, content, date_expires, date_created, date_updated
`

type CreateSnippetParams struct {
	SnippetID   string
	Title       pgtype.Text
	Content     pgtype.Text
	DateExpires pgtype.Timestamptz
	DateCreated pgtype.Timestamptz
	DateUpdated pgtype.Timestamptz
}

func (q *Queries) CreateSnippet(ctx context.Context, arg CreateSnippetParams) (Snippet, error) {
	row := q.db.QueryRow(ctx, createSnippet,
		arg.SnippetID,
		arg.Title,
		arg.Content,
		arg.DateExpires,
		arg.DateCreated,
		arg.DateUpdated,
	)
	var i Snippet
	err := row.Scan(
		&i.SnippetID,
		&i.Title,
		&i.Content,
		&i.DateExpires,
		&i.DateCreated,
		&i.DateUpdated,
	)
	return i, err
}

const deleteSnippet = `-- name: DeleteSnippet :exec
DELETE FROM snippets
  WHERE snippet_id = $1
`

func (q *Queries) DeleteSnippet(ctx context.Context, snippetID string) error {
	_, err := q.db.Exec(ctx, deleteSnippet, snippetID)
	return err
}

const getSnippet = `-- name: GetSnippet :one
SELECT snippet_id, title, content, date_expires, date_created, date_updated FROM snippets
  WHERE snippet_id = $1 LIMIT 1
`

func (q *Queries) GetSnippet(ctx context.Context, snippetID string) (Snippet, error) {
	row := q.db.QueryRow(ctx, getSnippet, snippetID)
	var i Snippet
	err := row.Scan(
		&i.SnippetID,
		&i.Title,
		&i.Content,
		&i.DateExpires,
		&i.DateCreated,
		&i.DateUpdated,
	)
	return i, err
}

const listLatestSnippets = `-- name: ListLatestSnippets :many
SELECT snippet_id, title, content, date_expires, date_created, date_updated FROM snippets
	WHERE date_expires > NOW()
	ORDER BY date_created DESC
	LIMIT 10
`

func (q *Queries) ListLatestSnippets(ctx context.Context) ([]Snippet, error) {
	rows, err := q.db.Query(ctx, listLatestSnippets)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Snippet
	for rows.Next() {
		var i Snippet
		if err := rows.Scan(
			&i.SnippetID,
			&i.Title,
			&i.Content,
			&i.DateExpires,
			&i.DateCreated,
			&i.DateUpdated,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listSnippets = `-- name: ListSnippets :many
SELECT snippet_id, title, content, date_expires, date_created, date_updated FROM snippets
  ORDER BY title
`

func (q *Queries) ListSnippets(ctx context.Context) ([]Snippet, error) {
	rows, err := q.db.Query(ctx, listSnippets)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Snippet
	for rows.Next() {
		var i Snippet
		if err := rows.Scan(
			&i.SnippetID,
			&i.Title,
			&i.Content,
			&i.DateExpires,
			&i.DateCreated,
			&i.DateUpdated,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateSnippet = `-- name: UpdateSnippet :exec
UPDATE snippets
  SET
    "title" = $2,
    "content" = $3,
    "date_expires" = $4,
    "date_updated" = $5
  WHERE snippet_id = $1
`

type UpdateSnippetParams struct {
	SnippetID   string
	Title       pgtype.Text
	Content     pgtype.Text
	DateExpires pgtype.Timestamptz
	DateUpdated pgtype.Timestamptz
}

func (q *Queries) UpdateSnippet(ctx context.Context, arg UpdateSnippetParams) error {
	_, err := q.db.Exec(ctx, updateSnippet,
		arg.SnippetID,
		arg.Title,
		arg.Content,
		arg.DateExpires,
		arg.DateUpdated,
	)
	return err
}
