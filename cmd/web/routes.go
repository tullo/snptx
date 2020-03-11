package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippet", app.showSnippet)
	mux.HandleFunc("/snippet/create", app.createSnippet)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Flow of control (reading from left to right):
	// recoverPanic ↔ logRequest ↔ secureHeaders ↔ servemux ↔ application handler
	return app.recoverPanic(app.logRequest(secureHeaders(mux)))
}
