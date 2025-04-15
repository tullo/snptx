-- name: GetUser :one
SELECT * FROM users
  WHERE user_id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
  ORDER BY name;

-- name: CreateUser :one
INSERT INTO users
	  (user_id, name, email, active, password_hash, roles, date_created, date_updated)
	VALUES
	  ($1, $2, $3, $4, $5, $6, $7, $8)
  RETURNING *;

-- name: UpdateUser :exec
UPDATE users
  SET
    "name" = $2,
    "email" = $3,
    "roles" = $4,
    "password_hash" = $5,
    "date_updated" = $6
  WHERE user_id = $1;

-- name: DeleteUser :exec
DELETE FROM users
  WHERE user_id = $1;
