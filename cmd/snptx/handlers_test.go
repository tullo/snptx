package main

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

	"github.com/tullo/snptx/internal/assert"
)

func TestPing(t *testing.T) {

	app := newTestApp(t)
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
	app := newTestApp(t)

	// start up a https test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid ID",
			urlPath:  "/snippet/view/1",
			wantCode: http.StatusOK,
			wantBody: "An old silent pond...",
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/snippet/view/2",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Negative ID",
			urlPath:  "/snippet/view/-1",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Decimal ID",
			urlPath:  "/snippet/view/1.23",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String ID",
			urlPath:  "/snippet/view/foo",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Empty ID",
			urlPath:  "/snippet/view/",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)

			assert.Equal(t, code, tt.wantCode)

			if tt.wantBody != "" {
				assert.StringContains(t, string(body), tt.wantBody)
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	app := newTestApp(t)

	// start up a https test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// make a GET /user/login request
	_, _, body := ts.get(t, "/user/login")

	// extract the CSRF token from the response body (html signup form)
	csrfToken := extractCSRFToken(t, string(body))

	tests := []struct {
		name         string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantBody     []byte
	}{
		{"Valid Submission", "alice@example.com", "validPa$$word", csrfToken, http.StatusSeeOther, nil},
		{"Empty Email", "", "validPa$$word", csrfToken, http.StatusUnprocessableEntity, []byte("This field cannot be blank")},
		{"Empty Password", "alice@example.com", "", csrfToken, http.StatusUnprocessableEntity, []byte("This field cannot be blank")},
		{"Invalid Password", "alice@example.com", "FooBarBaz", csrfToken, http.StatusUnprocessableEntity, []byte("Email or password is incorrect")},
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

func TestChangePassword(t *testing.T) {
	app := newTestApp(t)

	// start up a https test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// make a GET /user/login request
	_, _, body := ts.get(t, "/user/login")

	// extract the CSRF token from the response body (html login form)
	validCSRFToken := extractCSRFToken(t, string(body))

	// login with valid credentials
	form := url.Values{}
	form.Add("email", "alice@example.com")
	form.Add("password", "validPa$$word")
	form.Add("csrf_token", validCSRFToken)

	code, _, _ := ts.postForm(t, "/user/login", form)
	if code != http.StatusSeeOther {
		t.Errorf("want %d; got %d", http.StatusSeeOther, code)
	}

	/*
		u, _ := url.Parse(ts.URL + "/")
		for _, cookie := range ts.Client().Jar.Cookies(u) {
			fmt.Printf("  %s: %s\n", cookie.Name, cookie.Value)
		}
	*/

	tests := []struct {
		name                    string
		userEmail               string
		currentPassword         string
		newPassword             string
		newPasswordConfirmation string
		csrfToken               string
		wantCode                int
		wantBody                []byte
	}{
		{
			"Invalid CSRF Token",
			"",
			"",
			"",
			"",
			"wrongToken",
			http.StatusBadRequest,
			nil,
		},
		{
			"Blank Current Password",
			"alice@example.com",
			"",
			"someRandomString",
			"someRandomString",
			validCSRFToken,
			http.StatusUnprocessableEntity,
			[]byte("This field cannot be blank"),
		},
		{
			"Invalid Current Password",
			"alice@example.com",
			"GophersAreCute",
			"someRandomString",
			"someRandomString",
			validCSRFToken,
			http.StatusUnprocessableEntity,
			[]byte("Current password is incorrect"),
		},
		{
			"Invalid New Password 1",
			"alice@example.com",
			"validPa$$word",
			"gophers",
			"gophers",
			validCSRFToken,
			http.StatusUnprocessableEntity,
			[]byte("This field must be at least 10 characters long"),
		},
		{
			"Invalid New Password 2",
			"alice@example.com",
			"validPa$$word",
			"someRandomString",
			"gophers",
			validCSRFToken,
			http.StatusUnprocessableEntity,
			[]byte("This field must be at least 10 characters long"),
		},
		{
			"Invalid New Password 4",
			"alice@example.com",
			"validPa$$word",
			"someRandomString",
			"anotherRandomString",
			validCSRFToken,
			http.StatusUnprocessableEntity,
			[]byte("This field must be equal to the new password confirmation"),
		},
		{
			"Invalid New Password 5",
			"alice@example.com",
			"validPa$$word",
			"validPa$$word",
			"validPa$$word",
			validCSRFToken,
			http.StatusUnprocessableEntity,
			[]byte("This field cannot be equal to the current password"),
		},
		{
			"Valid Submission",
			"alice@example.com",
			"validPa$$word",
			"sup3rs3cr3t",
			"sup3rs3cr3t",
			validCSRFToken,
			http.StatusSeeOther, nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("email", tt.userEmail)
			form.Add("currentPassword", tt.currentPassword)
			form.Add("newPassword", tt.newPassword)
			form.Add("newPasswordConfirmation", tt.newPasswordConfirmation)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/user/change-password", form)
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
	app := newTestApp(t)

	// start up a https test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// make a GET /user/signup request
	_, _, body := ts.get(t, "/user/signup")

	// extract the CSRF token from the response body (html signup form)
	validCSRFToken := extractCSRFToken(t, string(body))

	//t.Log(validCSRFToken)

	const (
		validName     = "Bob"
		validPassword = "validPa$$word"
		validEmail    = "bob@example.com"
		formTag       = "<form action='/user/signup' method='POST' novalidate>"
	)

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantFormTag  string
	}{
		{
			name:         "Valid submission",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusSeeOther,
		},
		{
			name:         "Invalid CSRF Token",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    "wrongToken",
			wantCode:     http.StatusBadRequest,
		},
		{
			name:         "Empty name",
			userName:     "",
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty email",
			userName:     validName,
			userEmail:    "",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Invalid email",
			userName:     validName,
			userEmail:    "bob@example.",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Short password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "pa$$",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Duplicate email",
			userName:     validName,
			userEmail:    "dupe@example.com",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/user/signup", form)

			assert.Equal(t, code, tt.wantCode)

			if tt.wantFormTag != "" {
				assert.StringContains(t, string(body), tt.wantFormTag)
			}
		})
	}
}

// TestCreateSnippetForm checks that:
// - Unauthenticated users are redirected to the login form.
// - Authenticated users are shown the form to create a new snippet.
func TestCreateSnippetForm(t *testing.T) {

	app := newTestApp(t)

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
		csrfToken := extractCSRFToken(t, string(body))

		// post the form to login the user
		form := url.Values{}
		form.Add("email", "alice@example.com")
		form.Add("password", "validPa$$word")
		form.Add("csrf_token", csrfToken)
		login, _, _ := ts.postForm(t, "/user/login", form)

		if login != http.StatusSeeOther {
			t.Errorf("want %d; got %d", http.StatusSeeOther, login)
		}

		// authenticated users have access to the create snippet page
		code, _, body := ts.get(t, "/snippet/create")
		if code != http.StatusOK {
			t.Errorf("want %d; got %d", http.StatusOK, code)
		}

		// check that the authenticated user is shown the create snippet form
		formTag := "<form action='/snippet/create' method='POST'>"
		if !bytes.Contains(body, []byte(formTag)) {
			t.Errorf("want body %s to contain %q", body, formTag)
		}
	})
}

// TestDeleteSnippet checks that:
// - Unauthenticated users are redirected to the login form.
// - Authenticated users can delete snippets.
func TestDeleteSnippet(t *testing.T) {

	app := newTestApp(t)

	// start the https test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	t.Run("Unauthenticated", func(t *testing.T) {
		// get snippet page properties
		code, _, _ := ts.get(t, "/snippet/view/1")
		if code != http.StatusOK {
			t.Errorf("want %d; got %d", http.StatusOK, code)
		}

		// unauthenticated users may not delete snippets
		code, _, _ = ts.postForm(t, "/snippet/delete/1", nil)
		if code != http.StatusBadRequest {
			t.Errorf("want %d; got %d", http.StatusBadRequest, code)
		}
	})

	t.Run("Authenticated", func(t *testing.T) {

		// mimic the workflow of logging in as a user
		_, _, body := ts.get(t, "/user/login")

		// extract csrf token from the login page
		csrfToken := extractCSRFToken(t, string(body))

		// prepare login form
		form := url.Values{}
		form.Add("email", "alice@example.com")
		form.Add("password", "validPa$$word")
		form.Add("csrf_token", csrfToken)
		// login
		ts.postForm(t, "/user/login", form)

		// authenticated users may delete snippets
		// req, _ := http.NewRequest("DELETE", ts.URL+"/snippet/delete/1", nil)
		// res, err := ts.Client().Do(req)
		// if err != nil {
		// 	t.Fatal(err)
		// }
		// blob, _ := httputil.DumpResponse(res, true)
		// t.Log("xxxxx", string(blob))

		code, headers, _ := ts.postForm(t, "/snippet/delete/1", form)
		if code != http.StatusSeeOther {
			t.Errorf("want %d; got %d", http.StatusSeeOther, code)
		}
		// should be redirected to home page
		if headers.Get("Location") != "/" {
			t.Errorf("want %s; got %s", "/", headers.Get("Location"))
		}
	})
}
