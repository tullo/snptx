package schema

import "embed"

// Migrations contains the migrations needed to construct
// the database schema. Migration file pairs (up/down)
// should never be removed from this directory once they
// have been run in production.
//
//go:embed migrations
var migrations embed.FS
