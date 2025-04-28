package models

import (
	"github.com/alexedwards/scs/postgresstore"
	"github.com/tullo/snptx/internal/platform/database"
)

type SessionsStore struct {
	*postgresstore.PostgresStore
}

func NewSessionsStore(db *database.DB) SessionsStore {
	return SessionsStore{
		postgresstore.New(database.StdLibConnection(db.Pool)),
	}
}
