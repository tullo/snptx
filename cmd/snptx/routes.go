package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
	"github.com/tullo/snptx/ui"
)

func (a *app) routes() http.Handler {

	// 'standard' middleware used for every request
	// flow: recoverPanic ↔ logRequest ↔ secureHeaders
	standardMiddleware := alice.New(a.recoverPanic, a.logRequest, secureHeaders)

	// middleware specific to our dynamic application routes
	dynamicMiddleware := alice.New(a.session.Enable, noSurf, a.authenticate)

	mux := pat.New()
	mux.Get("/", dynamicMiddleware.ThenFunc(a.home))
	mux.Get("/about", dynamicMiddleware.ThenFunc(a.about))
	mux.Get("/snippet/create", dynamicMiddleware.Append(a.requireAuthentication).ThenFunc(a.createSnippetForm))
	mux.Post("/snippet/create", dynamicMiddleware.Append(a.requireAuthentication).ThenFunc(a.createSnippet))
	mux.Get("/snippet/:id", dynamicMiddleware.ThenFunc(a.showSnippet))
	mux.Post("/snippet/:id", dynamicMiddleware.Append(a.requireAuthentication).ThenFunc(a.deleteSnippet))
	mux.Get("/snippet/:id/edit", dynamicMiddleware.Append(a.requireAuthentication).ThenFunc(a.updateSnippetForm))
	mux.Post("/snippet/:id/edit", dynamicMiddleware.Append(a.requireAuthentication).ThenFunc(a.updateSnippet))

	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(a.signupUserForm))
	mux.Post("/user/signup", dynamicMiddleware.ThenFunc(a.signupUser))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(a.loginUserForm))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(a.loginUser))
	mux.Post("/user/logout", dynamicMiddleware.Append(a.requireAuthentication).ThenFunc(a.logoutUser))
	mux.Get("/user/profile", dynamicMiddleware.Append(a.requireAuthentication).ThenFunc(a.userProfile))
	mux.Get("/user/change-password", dynamicMiddleware.Append(a.requireAuthentication).ThenFunc(a.changePasswordForm))
	mux.Post("/user/change-password", dynamicMiddleware.Append(a.requireAuthentication).ThenFunc(a.changePassword))

	mux.Get("/ping", http.HandlerFunc(ping))
	mux.Get("/static/", http.FileServerFS(ui.Files))

	// Flow of control (reading from left to right):
	// standardMiddleware ↔ servemux ↔ dynamicMiddleware ↔ application handler
	return standardMiddleware.Then(mux)
}
