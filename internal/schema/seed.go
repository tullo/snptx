package schema

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Seed runs the set of seed-data queries against db. The queries are ran in a
// transaction and rolled back if any fail.
func Seed(ctx context.Context, db *pgxpool.Pool) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, seeds); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return err
		}
		return err
	}

	return tx.Commit(ctx)
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

	-- Create admin and regular User with password "goroutines"
	INSERT INTO users (user_id, name, email, roles, password_hash, date_created, date_updated) VALUES
		('405b059e-f6fc-4ed4-8532-d466264995e2', 'Admin Gopher', 'admin@example.com', '{ADMIN,USER}', '$argon2id$v=19$m=65536,t=1,p=1$k7s9K2Wa/mMakJbzH6C4IA$76puo7jSjAdaZwa6eOwYe6inF7bFDSDf/ryYdDAi8GI', '2020-09-20 00:00:00', '2020-09-20 00:00:00'),
		('9804845d-9b60-4177-880d-d15c431c36e2', 'User Gopher', 'user@example.com', '{USER}', '$argon2id$v=19$m=65536,t=1,p=1$uc7mAQY4Jbyd6xfw4IycWQ$7R6V5n/DENEg3m46HrCRFaVoooYKd5CYD+ZPSs3Ewg8', '2020-09-20 00:00:00', '2020-09-20 00:00:00')
		ON CONFLICT DO NOTHING;
`
