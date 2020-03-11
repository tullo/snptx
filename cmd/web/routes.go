package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {

	// 'standard' middleware used for every request
	// flow: recoverPanic ↔ logRequest ↔ secureHeaders
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippet", app.showSnippet)
	mux.HandleFunc("/snippet/create", app.createSnippet)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Flow of control (reading from left to right):
	// standardMiddleware ↔ servemux ↔ application handler
	return standardMiddleware.Then(mux)
}
