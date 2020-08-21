package main

import (
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"testing"
	"time"

	"github.com/golangcollege/sessions"
	"github.com/tullo/snptx/internal/mock"
)

// Capture the CSRF token value from the HTML for the user signup page
var csrfTokenRX = regexp.MustCompile(`<input type='hidden' name='csrf_token' value='(.+)'>`)

func extractCSRFToken(t *testing.T, body []byte) string {
	// extract the token from the HTML body
	matches := csrfTokenRX.FindSubmatch(body)
	// expecting an array with at least two entries (matched pattern & captured data)
	if len(matches) < 2 {
		t.Fatal("no csrf token found in body")
	}

	// unescape the rendered and html escaped base64 encoded string value
	return html.UnescapeString(string(matches[1]))
}

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

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// app struct instantiation using the mocks for the loggers and database models
	return &application{
		errorLog:      log.New(ioutil.Discard, "", 0),
		infoLog:       log.New(ioutil.Discard, "", 0),
		debug:         false,
		session:       session,
		shutdown:      shutdown,
		snippets:      &mock.SnippetModel{},
		templateCache: templateCache,
		users:         &mock.UserModel{},
		version:       "develop",
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

// get performs a GET request to a given url path on the test server
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
	// make a GET request against the test server
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

// postForm method for sending POST requests to the test server
func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header, []byte) {
	// make a POST request against the test server
	rs, err := ts.Client().PostForm(ts.URL+urlPath, form)
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
