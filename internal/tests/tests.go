package tests

import (
	"bytes"
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tullo/snptx/internal/platform/database"
	"github.com/tullo/snptx/internal/platform/database/databasetest"
	"github.com/tullo/snptx/internal/platform/web"
	"github.com/tullo/snptx/internal/schema"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

// ContainerSpec provides configuration for a docker container to run.
type ContainerSpec struct {
	Repository string
	Tag        string
	Port       string
	Args       []string
	Cmd        []string
}

func NewRoachDBSpec() ContainerSpec {
	return ContainerSpec{
		Repository: "cockroachdb/cockroach",
		Tag:        "v20.2.8",
		Port:       "26257/tcp",
		Cmd:        []string{"start-single-node", "--insecure", "--listen-addr=0.0.0.0"},
	}
}

func NewPostgresDBSpec() ContainerSpec {
	return ContainerSpec{
		Repository: "postgres",
		Tag:        "13.2-alpine",
		Port:       "5432",
		Args:       []string{"-e", "POSTGRES_USER=postgres", "-e", "POSTGRES_PASSWORD=postgres"},
	}
}

// NewUnit creates a test database inside a Docker container. It creates the
// required table structure but the database is otherwise empty.
//
// It does not return errors as this is intended for testing only.
// Instead it will call Fatal on the provided testing.T if anything goes wrong.
//
// It returns the database to use as well as a function to call at the end of
// the test.
func NewUnit(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	t.Log("waiting for database to be ready")

	p := NewPostgresDBSpec()

	img := bytes.NewBufferString(p.Repository)
	img.WriteByte(':')
	img.WriteString(p.Tag)

	c := databasetest.StartContainer(t, img.String(), p.Port, p.Args...)
	ctx := context.Background()
	cfg := database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.Host,
		Name:       "postgres",
		DisableTLS: true,
	}

	// Wait for the database to be ready. Wait 100ms longer between each attempt.
	// Do not try more than 20 times.
	var (
		db  *pgxpool.Pool
		err error
	)
	maxAttempts := 20
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		db, err = database.Connect(ctx, cfg)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
	}

	if err != nil {
		databasetest.DumpContainerLogs(t, c)
		databasetest.StopContainer(t, c)
		t.Fatalf("opening database connection: %v", err)
	}

	if err := schema.Migrate(database.ConnString(cfg)); err != nil {
		databasetest.StopContainer(t, c)
		t.Fatalf("migrating: %s", err)
	}

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown := func() {
		t.Helper()
		db.Close()
		databasetest.StopContainer(t, c)
	}

	return db, teardown
}

// Test owns state for running and shutting down tests.
type Test struct {
	DB      *pgxpool.Pool
	Log     *log.Logger
	t       *testing.T
	cleanup func()
}

// NewIntegration creates a database, seeds it, constructs an authenticator.
func NewIntegration(t *testing.T) *Test {
	t.Helper()

	// Initialize and seed database. Store the cleanup function call later.
	db, cleanup := NewUnit(t)

	deadline := time.Now().Add(time.Second * 15)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

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
