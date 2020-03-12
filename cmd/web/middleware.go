package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/tullo/snptx/pkg/models"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

// noSurf uses a customized CSRF cookie with the Secure, Path and HttpOnly flags set
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}

// recoverPanic recovers the panic and logs the cause
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// called last on the way up in the middleware chain while Go unwinds the stack
		defer func() {
			// check if there has been a panic or not
			if err := recover(); err != nil {
				// Trigger the Go server to automatically close the current connection
				// after a response has been sent.
				w.Header().Set("Connection", "close")
				// format error with default textual representation
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// requireAuthentication redirects the unauthenticated user to the login page
func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		// pages that require authentication should not be stored in caches
		// (browser cache or other intermediary cache)
		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

// authenticate checks the database for user status (active)
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if user is logged in
		exists := app.session.Exists(r, "authenticatedUserID")
		if !exists {
			next.ServeHTTP(w, r)
			return
		}

		// do a database lookup with user ID from session data
		user, err := app.users.Get(app.session.GetInt(r, "authenticatedUserID"))
		// remove user ID from session if user not found or deactivated
		if errors.Is(err, models.ErrNoRecord) || !user.Active {
			app.session.Remove(r, "authenticatedUserID")
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			app.serverError(w, err)
			return
		}

		// request is coming from an authenticated & 'active' user,
		// add key/value pair to the request context - to be used further down the chain
		ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
