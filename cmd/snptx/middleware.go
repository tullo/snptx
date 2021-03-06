package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/tullo/snptx/internal/user"

	"github.com/justinas/nosurf"
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

func (a *app) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.log.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}

// recoverPanic recovers the panic and logs the cause
func (a *app) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// called last on the way up in the middleware chain while Go unwinds the stack
		defer func() {
			// check if there has been a panic or not
			if err := recover(); err != nil {
				// Trigger the Go server to automatically close the current connection
				// after a response has been sent.
				w.Header().Set("Connection", "close")
				// format error with default textual representation
				a.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// requireAuthentication redirects the unauthenticated user to the login page
func (a *app) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.isAuthenticated(r) {
			// add the path the user is trying to access to session data
			a.session.Put(r, "redirectPathAfterLogin", r.URL.Path)
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
func (a *app) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if user is logged in
		exists := a.session.Exists(r, "authenticatedUserID")
		if !exists {
			next.ServeHTTP(w, r)
			return
		}

		uid := a.session.GetString(r, "authenticatedUserID")
		if len(uid) == 0 {
			// clean up user session state
			a.session.Remove(r, "authenticatedUserID")
			next.ServeHTTP(w, r)
			return
		}

		// lookup user with id from users session data
		usr, err := a.users.QueryByID(r.Context(), uid)
		if errors.Is(err, user.ErrNotFound) {
			// remove session key if no records found
			a.session.Remove(r, "authenticatedUserID")
			next.ServeHTTP(w, r)
			return
		}

		if usr != nil && !usr.Active {
			// remove session key if user record is in deactivated state
			a.session.Remove(r, "authenticatedUserID")
			next.ServeHTTP(w, r)
			return
		}

		if err != nil {
			a.serverError(w, err)
			return
		}

		// request is coming from an authenticated & 'active' user,
		// add key/value pair to the request context - to be used further down the chain
		ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
