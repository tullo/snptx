package main

import (
	"net/http"

	"github.com/justinas/alice"
	"github.com/tullo/snptx/ui"
)

func (a *app) routes() http.Handler {

	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	mux.HandleFunc("GET /ping", ping)
	// mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "static/img/favicon.ico")
	// })

	// middleware specific to our dynamic application routes
	dynamic := alice.New(a.sessionManager.LoadAndSave, noSurf, a.authenticate)

	mux.Handle("GET /{$}", dynamic.ThenFunc(a.home))
	mux.Handle("GET /about", dynamic.ThenFunc(a.about))

	mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(a.snippetView))

	mux.Handle("GET /user/signup", dynamic.ThenFunc(a.userSignupForm))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(a.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(a.loginUserForm))
	mux.Handle("POST /user/login", dynamic.ThenFunc(a.userLoginPost))

	protected := dynamic.Append(a.requireAuthentication)

	mux.Handle("GET /snippet/create", protected.ThenFunc(a.snippetCreateForm))
	mux.Handle("POST /snippet/create", protected.ThenFunc(a.snippetCreatePost))
	mux.Handle("GET /snippet/edit/{id}", protected.ThenFunc(a.updateSnippetForm))
	mux.Handle("POST /snippet/edit/{id}", protected.ThenFunc(a.updateSnippetPost))
	mux.Handle("POST /snippet/delete/{id}", protected.ThenFunc(a.snippetDeletePost))

	mux.Handle("GET /user/change-password", protected.ThenFunc(a.changePasswordForm))
	mux.Handle("POST /user/change-password", protected.ThenFunc(a.changePasswordPost))
	mux.Handle("POST /user/logout", protected.ThenFunc(a.logoutUserPost))
	mux.Handle("GET /user/profile", protected.ThenFunc(a.userProfile))

	// 'standard' middleware used for every request
	// flow of control: recoverPanic ↔ logRequest ↔ commonHeaders
	standard := alice.New(a.recoverPanic, a.logRequest, commonHeaders)

	// Flow of control (reading from left to right):
	// standard ↔ servemux ↔ dynamic ↔ application handler
	return standard.Then(mux)
}
