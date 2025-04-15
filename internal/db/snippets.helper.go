package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func GetCreateSnippetParams(id, title, content string, exp, create, up time.Time) CreateSnippetParams {
	return CreateSnippetParams{
		SnippetID:   id,
		Title:       pgtype.Text{String: title, Valid: true},
		Content:     pgtype.Text{String: content, Valid: true},
		DateExpires: pgtype.Timestamptz{Time: exp, Valid: true},
		DateCreated: pgtype.Timestamptz{Time: create, Valid: true},
		DateUpdated: pgtype.Timestamptz{Time: up, Valid: true},
	}

}

func GetUpdateSnippetParams(id, title, content string, exp, up time.Time) UpdateSnippetParams {
	return UpdateSnippetParams{
		SnippetID:   id,
		Title:       pgtype.Text{String: title, Valid: true},
		Content:     pgtype.Text{String: content, Valid: true},
		DateExpires: pgtype.Timestamptz{Time: exp, Valid: true},
		DateUpdated: pgtype.Timestamptz{Time: up, Valid: true},
	}
}
