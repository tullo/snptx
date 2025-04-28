package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*pgxpool.Pool
}

// Config is the required properties to use the database.
type Config struct {
	User       string
	Password   string
	Host       string
	Name       string
	DisableTLS bool
}

func StdLibConnection(p *pgxpool.Pool) *sql.DB {
	return sql.OpenDB(stdlib.GetConnector(*p.Config().ConnConfig))
}

// Connect establishes a database connection based on the configuration.
func Connect(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dbaddr, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Println("DATABASE_URL env var not defined")
	}

	p, err := pgxpool.New(ctx, dbaddr)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}

	return p, nil
}

// ConnString translates config to a db connection string.
func ConnString(cfg Config) string {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	// Query parameters.
	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	return u.String()
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *pgxpool.Pool) error {
	// Run a simple query to determine connectivity. The db has a "Ping" method
	// but it can false-positive when it was previously able to talk to the
	// database but the database has since gone away. Running this query forces a
	// round trip to the database.
	const q = `SELECT true`
	var tmp bool
	return pgxscan.Get(ctx, db, &tmp, q)
}

// SanitizeDatabaseName ensures that the database name is a valid postgres identifier.
func SanitizeDatabaseName(schema string) string {
	return pgx.Identifier{schema}.Sanitize()
}

// ConnstrWithDatabase changes the main database in the connection string.
func ConnstrWithDatabase(connstr, database string) (string, error) {
	u, err := url.Parse(connstr)
	if err != nil {
		return "", fmt.Errorf("invalid connstr: %q", connstr)
	}
	u.Path = database
	return u.String(), nil
}

func ConnectWithURI(ctx context.Context, uri string) (*DB, error) {
	pool, err := pgxpool.New(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}

	conf, err := traceLogConfig(pool)
	if err != nil {
		return nil, fmt.Errorf("database config error: %w", err)
	}

	pool, err = pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}

	db := DB{pool}

	return &db, nil
}

func traceLogConfig(pool *pgxpool.Pool) (*pgxpool.Config, error) {
	// logger, err := zap.NewDevelopmentConfig().Build()
	// if err != nil {
	// 	return nil, fmt.Errorf("zap logger error: %w", err)
	// }
	conf := pool.Config()
	// conf.ConnConfig.Tracer = &tracelog.TraceLog{
	// 	Logger:   zapadapter.NewLogger(logger),
	// 	LogLevel: tracelog.LogLevelDebug,
	// }

	return conf, nil

}
