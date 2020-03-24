package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/justinas/nosurf"
)

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// go one step back in the stack trace to get the file name and line number
	app.errorLog.Output(2, trace)

	// when running in debug mode,
	// write detailed errors and stack traces to the http response
	if app.debug {
		http.Error(w, trace, http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

	app.SignalShutdown()
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}
	td.CurrentYear = time.Now().Year()

	// add CSRF token to the template data
	td.CSRFToken = nosurf.Token(r)

	// retrieve the value for the flash key and delete the key in one step
	// add flash message to the template data
	td.Flash = app.session.PopString(r, "flash")

	// add authentication status to the template data
	td.IsAuthenticated = app.isAuthenticated(r)

	return td
}

// isAuthenticated checks if the request is from an authenticated user
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(contextKeyIsAuthenticated).(bool)
	if !ok {
		// key not found in ctx, or value was not a boolean
		return false
	}
	return isAuthenticated
}

func (app *application) render(w http.ResponseWriter, r *http.Request, name string, data *templateData) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// stage 1: write template into buffer
	buf := new(bytes.Buffer)
	err := ts.Execute(buf, app.addDefaultData(data, r))
	if err != nil {
		app.serverError(w, err)
		return
	}

	// stage 2: write rendered content
	buf.WriteTo(w)
}
