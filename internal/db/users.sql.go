// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: users.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const changePassword = `-- name: ChangePassword :exec
UPDATE users
  SET
    "password_hash" = $1
  WHERE
    "user_id" = $2
`

type ChangePasswordParams struct {
	PasswordHash pgtype.Text
	UserID       string
}

func (q *Queries) ChangePassword(ctx context.Context, arg ChangePasswordParams) error {
	_, err := q.db.Exec(ctx, changePassword, arg.PasswordHash, arg.UserID)
	return err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users
	  (user_id, name, email, active, password_hash, roles, date_created, date_updated)
	VALUES
	  ($1, $2, $3, $4, $5, $6, $7, $8)
  RETURNING user_id, name, email, active, roles, password_hash, date_created, date_updated
`

type CreateUserParams struct {
	UserID       string
	Name         pgtype.Text
	Email        pgtype.Text
	Active       pgtype.Bool
	PasswordHash pgtype.Text
	Roles        []string
	DateCreated  pgtype.Timestamptz
	DateUpdated  pgtype.Timestamptz
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.UserID,
		arg.Name,
		arg.Email,
		arg.Active,
		arg.PasswordHash,
		arg.Roles,
		arg.DateCreated,
		arg.DateUpdated,
	)
	var i User
	err := row.Scan(
		&i.UserID,
		&i.Name,
		&i.Email,
		&i.Active,
		&i.Roles,
		&i.PasswordHash,
		&i.DateCreated,
		&i.DateUpdated,
	)
	return i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users
  WHERE
    "user_id" = $1
`

func (q *Queries) DeleteUser(ctx context.Context, userID string) error {
	_, err := q.db.Exec(ctx, deleteUser, userID)
	return err
}

const getUser = `-- name: GetUser :one
SELECT user_id, name, email, active, roles, password_hash, date_created, date_updated FROM users
  WHERE "user_id" = $1
`

func (q *Queries) GetUser(ctx context.Context, userID string) (User, error) {
	row := q.db.QueryRow(ctx, getUser, userID)
	var i User
	err := row.Scan(
		&i.UserID,
		&i.Name,
		&i.Email,
		&i.Active,
		&i.Roles,
		&i.PasswordHash,
		&i.DateCreated,
		&i.DateUpdated,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT user_id, name, email, active, roles, password_hash, date_created, date_updated FROM users
  WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email pgtype.Text) (User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.UserID,
		&i.Name,
		&i.Email,
		&i.Active,
		&i.Roles,
		&i.PasswordHash,
		&i.DateCreated,
		&i.DateUpdated,
	)
	return i, err
}

const listUsers = `-- name: ListUsers :many
SELECT user_id, name, email, active, roles, password_hash, date_created, date_updated FROM users
  ORDER BY name
`

func (q *Queries) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := q.db.Query(ctx, listUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.UserID,
			&i.Name,
			&i.Email,
			&i.Active,
			&i.Roles,
			&i.PasswordHash,
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

const updateUser = `-- name: UpdateUser :exec
UPDATE users
  SET
    "name" = $2,
    "email" = $3,
    "roles" = $4,
    "password_hash" = $5,
    "date_updated" = $6
  WHERE user_id = $1
`

type UpdateUserParams struct {
	UserID       string
	Name         pgtype.Text
	Email        pgtype.Text
	Roles        []string
	PasswordHash pgtype.Text
	DateUpdated  pgtype.Timestamptz
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) error {
	_, err := q.db.Exec(ctx, updateUser,
		arg.UserID,
		arg.Name,
		arg.Email,
		arg.Roles,
		arg.PasswordHash,
		arg.DateUpdated,
	)
	return err
}
