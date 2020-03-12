package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golangcollege/sessions"
	"github.com/tullo/snptx/pkg/models/mock"
)

// newTestApplication creates an application struct with mock loggers
func newTestApplication(t *testing.T) *application {
	// initialize template cache
	templateCache, err := newTemplateCache("./../../ui/html/")
	if err != nil {
		t.Fatal(err)
	}

	// session manager instance that mirrors production settings
	// sample generation of secret bytes 'openssl rand -base64 32'
	session := sessions.New([]byte("zBtjT1J8wWrvUCuEZf+YbBa41nKYlCKiNLeS5AGdmiQ="))
	// sessions expire after 12 hours
	session.Lifetime = 12 * time.Hour
	// set the secure flag on session cookies
	session.Secure = true
	// mitigate cross site request forgry csrf
	session.SameSite = http.SameSiteStrictMode

	// app struct instantiation using the mocks for the loggers and database models
	return &application{
		errorLog:      log.New(ioutil.Discard, "", 0),
		infoLog:       log.New(ioutil.Discard, "", 0),
		session:       session,
		snippets:      &mock.SnippetModel{},
		templateCache: templateCache,
		users:         &mock.UserModel{},
	}
}

type testServer struct {
	*httptest.Server
}

// newTestServer initalizes and returns a new instance of testServer
func newTestServer(t *testing.T, h http.Handler) *testServer {

	// spinup a https server for the duration of the test
	ts := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	// add the cookie jar to the client, so that response cookies are stored
	// and then sent with subsequent requests
	ts.Client().Jar = jar

	// disabling the default behaviour for redirect-following for the client
	// returning the error forces it to immediately return the received response
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

// get performs a get request to a given url path on the test server
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
	// make a get request against the test server
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()

	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	return rs.StatusCode, rs.Header, body
}
