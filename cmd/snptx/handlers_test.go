package main

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
)

func TestPing(t *testing.T) {

	app := newTestApplication(t)
	// start up a https test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// make a GET /ping request
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

func TestShowSnippet(t *testing.T) {
	app := newTestApplication(t)

	// start up a https test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody []byte
	}{
		{"Valid ID", "/snippet/1", http.StatusOK, []byte("An old silent pond...")},
		{"Non-existent ID", "/snippet/2", http.StatusNotFound, nil},
		{"Negative ID", "/snippet/-1", http.StatusNotFound, nil},
		{"Decimal ID", "/snippet/1.23", http.StatusNotFound, nil},
		{"String ID", "/snippet/foo", http.StatusNotFound, nil},
		{"Empty ID", "/snippet/", http.StatusNotFound, nil},
		{"Trailing slash", "/snippet/1/", http.StatusNotFound, nil},
		{"Internal Error", "/snippet/66", http.StatusInternalServerError, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body to contain %q", tt.wantBody)
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	app := newTestApplication(t)

	// start up a https test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// make a GET /user/login request
	_, _, body := ts.get(t, "/user/login")

	// extract the CSRF token from the response body (html signup form)
	csrfToken := extractCSRFToken(t, body)

	tests := []struct {
		name         string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantBody     []byte
	}{
		{"Valid submission", "user@example.com", "gophers", csrfToken, http.StatusOK, nil},
		{"Empty email", "", "gophers", csrfToken, http.StatusOK, []byte("Email or Password is incorrect")},
		{"Empty password", "user@example.com", "", csrfToken, http.StatusOK, []byte("Email or Password is incorrect")},
		{"Invalid password", "user@example.com", "FooBarBaz", csrfToken, http.StatusOK, []byte("Email or Password is incorrect")},
		{"Invalid CSRF Token", "", "", "wrongToken", http.StatusBadRequest, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/user/login", form)

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body %s to contain %q", body, tt.wantBody)
			}
		})
	}
}

func TestSignupUser(t *testing.T) {
	app := newTestApplication(t)

	// start up a https test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// make a GET /user/signup request
	_, _, body := ts.get(t, "/user/signup")

	// extract the CSRF token from the response body (html signup form)
	csrfToken := extractCSRFToken(t, body)

	//t.Log(csrfToken)

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantBody     []byte
	}{
		{"Valid submission", "Bob", "bob@example.com", "validPa$$word", csrfToken, http.StatusSeeOther, nil},
		{"Empty name", "", "bob@example.com", "validPa$$word", csrfToken, http.StatusOK, []byte("This field cannot be blank")},
		{"Empty email", "Bob", "", "validPa$$word", csrfToken, http.StatusOK, []byte("This field cannot be blank")},
		{"Empty password", "Bob", "bob@example.com", "", csrfToken, http.StatusOK, []byte("This field cannot be blank")},
		{"Invalid email (incomplete domain)", "Bob", "bob@example.", "validPa$$word", csrfToken, http.StatusOK, []byte("This field is invalid")},
		{"Invalid email (missing @)", "Bob", "bobexample.com", "validPa$$word", csrfToken, http.StatusOK, []byte("This field is invalid")},
		{"Invalid email (missing local part)", "Bob", "@example.com", "validPa$$word", csrfToken, http.StatusOK, []byte("This field is invalid")},
		{"Short password", "Bob", "bob@example.com", "pa$$word", csrfToken, http.StatusOK, []byte("This field is too short (minimum is 10 characters)")},
		{"Duplicate email", "Bob", "dupe@example.com", "validPa$$word", csrfToken, http.StatusOK, []byte("Address is already in use")},
		{"Invalid CSRF Token", "", "", "", "wrongToken", http.StatusBadRequest, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/user/signup", form)

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body %s to contain %q", body, tt.wantBody)
			}
		})
	}
}

// TestCreateSnippetForm checks that:
// - Unauthenticated users are redirected to the login form.
// - Authenticated users are shown the form to create a new snippet.
func TestCreateSnippetForm(t *testing.T) {

	app := newTestApplication(t)

	// start the https test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// run sub-test
	t.Run("Unauthenticated", func(t *testing.T) {
		code, headers, _ := ts.get(t, "/snippet/create")
		if code != http.StatusSeeOther {
			t.Errorf("want %d; got %d", http.StatusSeeOther, code)
		}
		if headers.Get("Location") != "/user/login" {
			t.Errorf("want %s; got %s", "/user/login", headers.Get("Location"))
		}
	})

	// run sub-test
	t.Run("Authenticated", func(t *testing.T) {

		// mimic the workflow of logging in as a user to authenticate

		// the get call returns: code, headers, body_bytes
		_, _, body := ts.get(t, "/user/login")

		// extract csrf token from the page with the login form
		csrfToken := extractCSRFToken(t, body)

		// post the form to login the user
		form := url.Values{}
		form.Add("email", "alice@example.com")
		form.Add("password", "")
		form.Add("csrf_token", csrfToken)
		ts.postForm(t, "/user/login", form)

		// the authenticated user have access to the create snippet page
		code, _, body := ts.get(t, "/snippet/create")
		if code != http.StatusOK {
			t.Errorf("want %d; got %d", http.StatusOK, code)
		}
		//t.Log(string(body))

		// check that the authenticated user is shown the create snippet form
		formTag := "<form action='/snippet/create' method='POST'>"
		if !bytes.Contains(body, []byte(formTag)) {
			t.Errorf("want body %s to contain %q", body, formTag)
		}
	})
}
