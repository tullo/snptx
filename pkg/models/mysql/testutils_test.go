package mysql

import (
	"database/sql"
	"io/ioutil"
	"testing"
)

// newTestDB establishes a sql.DB connection pool for the test database
func newTestDB(t *testing.T) (*sql.DB, func()) {

	// configure database driver to support executing multiple SQL statements for db.Exec() calls
	db, err := sql.Open("mysql", "test_web:pass@/test_snptx?parseTime=true&multiStatements=true")
	if err != nil {
		t.Fatal(err)
	}

	// read the setup SQL script
	script, err := ioutil.ReadFile("./testdata/setup.sql")
	if err != nil {
		t.Fatal(err)
	}
	// execute the statements
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	// return the connection pool and an anonymous function
	return db, func() {
		// read the teardown SQL script
		script, err := ioutil.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}
		// execute the statements
		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}

		db.Close()
	}
}
