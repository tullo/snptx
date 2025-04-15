package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func GetCreateUserParams(id, name, email string, active bool,
	hash string, roles []string, create, up time.Time) CreateUserParams {
	return CreateUserParams{
		UserID:       id,
		Name:         pgtype.Text{String: name, Valid: true},
		Email:        pgtype.Text{String: email, Valid: true},
		Active:       pgtype.Bool{Bool: active, Valid: true},
		Roles:        roles,
		PasswordHash: pgtype.Text{String: hash, Valid: true},
		DateCreated:  pgtype.Timestamptz{Time: create, Valid: true},
		DateUpdated:  pgtype.Timestamptz{Time: up, Valid: true},
	}

}

func GetUpdateUserParams(id, name, email string, roles []string,
	hash string, up time.Time) UpdateUserParams {
	return UpdateUserParams{
		UserID:       id,
		Name:         pgtype.Text{String: name, Valid: true},
		Email:        pgtype.Text{String: email, Valid: true},
		Roles:        roles,
		PasswordHash: pgtype.Text{String: hash, Valid: true},
		DateUpdated:  pgtype.Timestamptz{Time: up, Valid: true},
	}
}

func AsText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func GetChangePasswordParams(hash, uid string) ChangePasswordParams {
	return ChangePasswordParams{
		PasswordHash: pgtype.Text{String: hash, Valid: true},
		UserID:       uid,
	}
}
