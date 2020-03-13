package schema

import (
	"github.com/dimiro1/darwin"
	"github.com/jmoiron/sqlx"
)

// Migrate attempts to bring the schema for db up to date with the migrations
// defined in this package.
func Migrate(db *sqlx.DB) error {
	driver := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})

	d := darwin.New(driver, migrations, nil)

	return d.Migrate()
}

// migrations contains the queries needed to construct the database schema.
var migrations = []darwin.Migration{
	{
		Version:     1,
		Description: "Add snippets",
		Script: `
CREATE TABLE snippets (
	snippet_id INTEGER,
	title VARCHAR(100) NOT NULL,
	content TEXT NOT NULL,
	created TIMESTAMP NOT NULL,
	expires TIMESTAMP NOT NULL,
	PRIMARY KEY (snippet_id)
);`,
	},
	{
		Version:     2,
		Description: "Add idx snippets(created)",
		Script:      `CREATE INDEX idx_snippets_created ON snippets(created);`,
	},
	{
		Version:     3,
		Description: "Add idx snippets(created)",
		Script: `
CREATE TABLE users (
	user_id       INTEGER,
	name          TEXT,
	email         TEXT UNIQUE,
	password_hash TEXT,
	created TIMESTAMP,
	active BOOLEAN NOT NULL DEFAULT TRUE,
	PRIMARY KEY (user_id)
);`,
	},
}
