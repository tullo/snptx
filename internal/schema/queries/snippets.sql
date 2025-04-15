-- name: GetSnippet :one
SELECT * FROM snippets
  WHERE snippet_id = $1 LIMIT 1;

-- name: ListSnippets :many
SELECT * FROM snippets
  ORDER BY title;

-- name: ListLatestSnippets :many
SELECT * FROM snippets
	WHERE date_expires > NOW()
	ORDER BY date_created DESC
	LIMIT 10;

-- name: CreateSnippet :one
INSERT INTO snippets
  (snippet_id, title, content, date_expires, date_created, date_updated)
  VALUES
    ($1, $2, $3, $4, $5, $6)
  RETURNING *;

-- name: UpdateSnippet :exec
UPDATE snippets
  SET
    "title" = $2,
    "content" = $3,
    "date_expires" = $4,
    "date_updated" = $5
  WHERE snippet_id = $1;

-- name: DeleteSnippet :exec
DELETE FROM snippets
  WHERE snippet_id = $1;
