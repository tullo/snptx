package schema

import (
	"github.com/jmoiron/sqlx"
)

// Seed runs the set of seed-data queries against db. The queries are ran in a
// transaction and rolled back if any fail.
func Seed(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(seeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// seeds is a string constant containing all of the queries needed to get the
// db seeded to a useful state for development.
// escape string syntax (E'...') https://www.postgresql.org/docs/12/runtime-config-compatible.html
const seeds = `
	INSERT INTO snippets (snippet_id, title, content, date_created, date_updated, date_expires) VALUES
		('a2b0639f-2cc6-44b8-b97b-15d69dbb511e', 'An old silent pond',
		E'A frog jumps into the pond,\nsplash! Silence again.\n\n– Matsuo Bashō',
		NOW(), NOW(), NOW() + INTERVAL '365 days'),
		('72f8b983-3eb4-48db-9ed0-e45cc6bd716b', 'Over the wintry forest',
		E'Over the wintry\nforest, winds howl in rage\nwith no leaves to blow.\n\n– Natsume Soseki',
		NOW(), NOW(), NOW() + INTERVAL '365 days'),
		('98b6d4b8-f04b-4c79-8c2e-a0aef46854b7', 'Haiku',
		E'First autumn morning:\n\nthe mirror I stare into\nshows my father''s face.\n\n– Murakami Kijo',
		NOW(), NOW(), NOW() + INTERVAL '7 days')
		ON CONFLICT DO NOTHING;

	-- Create admin and regular User with password "gophers"
	INSERT INTO users (user_id, name, email, roles, password_hash, date_created, date_updated) VALUES
		('5cf37266-3473-4006-984f-9325122678b7', 'Admin Gopher', 'admin@example.com', '{ADMIN,USER}', '$2a$10$1ggfMVZV6Js0ybvJufLRUOWHS5f6KneuP0XwwHpJ8L8ipdry9f2/a', '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
		('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'User Gopher', 'user@example.com', '{USER}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', '2019-03-24 00:00:00', '2019-03-24 00:00:00')
		ON CONFLICT DO NOTHING;
`
