package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/justinas/nosurf"
)

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (a *app) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// go one step back in the stack trace to get the file name and line number
	a.errorLog.Output(2, trace)

	// when running in debug mode,
	// write detailed errors and stack traces to the http response
	if a.debug {
		http.Error(w, trace, http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

	a.SignalShutdown()
}

func (a *app) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (a *app) notFound(w http.ResponseWriter) {
	a.clientError(w, http.StatusNotFound)
}

func (a *app) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}
	td.CurrentYear = time.Now().Year()
	td.Version = a.version[:7]

	// add CSRF token to the template data
	td.CSRFToken = nosurf.Token(r)

	// retrieve the value for the flash key and delete the key in one step
	// add flash message to the template data
	td.Flash = a.session.PopString(r, "flash")

	// add authentication status to the template data
	td.IsAuthenticated = a.isAuthenticated(r)

	return td
}

// isAuthenticated checks if the request is from an authenticated user
func (a *app) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(contextKeyIsAuthenticated).(bool)
	if !ok {
		// key not found in ctx, or value was not a boolean
		return false
	}
	return isAuthenticated
}

func (a *app) render(w http.ResponseWriter, r *http.Request, name string, data *templateData) {
	ts, ok := a.templateCache[name]
	if !ok {
		a.serverError(w, fmt.Errorf("the template %s does not exist", name))
		return
	}

	// stage 1: write template into buffer
	buf := new(bytes.Buffer)
	err := ts.Execute(buf, a.addDefaultData(data, r))
	if err != nil {
		log.Printf("render err=%v+\n", err)
		a.serverError(w, err)
		return
	}

	// stage 2: write rendered content
	buf.WriteTo(w)
}
