package main

import (
	"net/http"
	"testing"
)

func TestPing(t *testing.T) {

	app := newTestApplication(t)
	// start up a https test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// make a get /ping request
	code, _, body := ts.get(t, "/ping")

	// check the status code written by the handler
	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}

	// check the response body
	if string(body) != "OK" {
		t.Errorf("want body to equal %q", "OK")
	}
}
