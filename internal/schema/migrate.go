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
			snippet_id 	  UUID,
			title 		  TEXT,
			content		  TEXT,
			date_expires  TIMESTAMP WITH TIME ZONE,
			date_created  TIMESTAMP WITH TIME ZONE,
			date_updated  TIMESTAMP WITH TIME ZONE,
			PRIMARY KEY (snippet_id)
		);`,
	},
	{
		Version:     2,
		Description: "Add idx snippets(date_created)",
		Script:      `CREATE INDEX idx_snippets_created ON snippets(date_created);`,
	},
	{
		Version:     3,
		Description: "Add users",
		Script: `
		CREATE TABLE users (
			user_id       UUID,
			name          TEXT,
			email         TEXT UNIQUE,
			roles         TEXT[],
			password_hash TEXT,
			date_created  TIMESTAMP WITH TIME ZONE,
			date_updated  TIMESTAMP WITH TIME ZONE,
			PRIMARY KEY (user_id)
		);`,
	},
}
