package tests

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tullo/snptx/internal/platform/database"
	"github.com/tullo/snptx/internal/platform/web"
	"github.com/tullo/snptx/internal/schema"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

func createDatabase(ctx context.Context, pool *database.DB, name string) error {
	_, err := pool.Exec(ctx, `CREATE DATABASE `+database.SanitizeDatabaseName(name)+`;`)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
	}

	return err
}

// dropDatabase drops the specific database.
func dropDatabase(ctx context.Context, pool *database.DB, name string) error {
	_, err := pool.Exec(ctx, `DROP DATABASE `+database.SanitizeDatabaseName(name)+`;`)
	return err
}

// NewUnit creates a test database inside a Docker container. It creates the
// required table structure but the database is otherwise empty.
//
// It does not return errors as this is intended for testing only.
// Instead it will call Fatal on the provided testing.T if anything goes wrong.
//
// It returns the database to use as well as a function to call at the end of
// the test.
func NewUnit(t *testing.T, ctx context.Context) (*database.DB, func()) {
	t.Helper()

	t.Log("waiting for database to be ready")

	dbaddr, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		t.Fatal("database url not defined")
	}
	maindb, err := database.ConnectWithURI(ctx, dbaddr)
	if err != nil {
		t.Fatal(fmt.Errorf("database connection error: %w", err))
	}

	// Create a unique database name so that our parallel tests don't clash.
	var id [8]byte
	rand.Read(id[:])
	uniqueName := t.Name() + "/" + hex.EncodeToString(id[:])

	err = createDatabase(ctx, maindb, uniqueName)
	if err != nil {
		t.Fatal(fmt.Errorf("database creation error: %w", err))
	}

	// Modify the connection string to use a different database.
	connstr, err := database.ConnstrWithDatabase(dbaddr, uniqueName)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to modify connnection string: %w", err))
	}

	db, err := database.ConnectWithURI(ctx, connstr)
	if err != nil {
		t.Fatal(fmt.Errorf("database connection error: %w", err))
	}

	err = schema.Migrate(connstr)
	if err != nil {
		t.Fatal(err)
	}

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown := func() {
		t.Helper()
		err = dropDatabase(context.TODO(), maindb, uniqueName)
		if err != nil {
			t.Fatal(fmt.Errorf("database deletetion error: %w", err))
		}
		db.Close()
	}

	return db, teardown
}

// Test owns state for running and shutting down tests.
type Test struct {
	DB      *database.DB
	Log     *log.Logger
	t       *testing.T
	cleanup func()
}

// NewIntegration creates a database, seeds it, constructs an authenticator.
func NewIntegration(t *testing.T) *Test {
	t.Helper()

	deadline := time.Now().Add(time.Second * 15)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// Initialize and seed database. Store the cleanup function call later.
	db, cleanup := NewUnit(t, ctx)

	if err := schema.Seed(ctx, db); err != nil {
		t.Fatal(err)
	}

	// Create the logger to use.
	logger := log.New(os.Stdout, "TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	return &Test{
		DB:      db,
		Log:     logger,
		t:       t,
		cleanup: cleanup,
	}
}

// Teardown releases any resources used for the test.
func (test *Test) Teardown() {
	test.cleanup()
}

// Context returns an app level context for testing.
func Context() context.Context {
	values := web.Values{
		TraceID: uuid.New().String(),
		Now:     time.Now(),
	}

	return context.WithValue(context.Background(), web.KeyValues, &values)
}

// StringPointer is a helper to get a *string from a string. It is in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func StringPointer(s string) *string {
	return &s
}

// IntPointer is a helper to get a *int from a int. It is in the tests package
// because we normally don't want to deal with pointers to basic types but it's
// useful in some tests.
func IntPointer(i int) *int {
	return &i
}

// TimePointer is a helper to get a *time from a time.
func TimePointer(t time.Time) *time.Time {
	return &t
}
