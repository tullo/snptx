package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
)

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (a *app) serverError(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// go one step back in the stack trace to get the file name and line number
	a.log.Output(2, trace)

	// when running in debug mode,
	// write detailed errors and stack traces to the http response
	if a.debug {
		http.Error(w, trace, http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

	// a.SignalShutdown() TODO
}

func (a *app) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (a *app) notFound(w http.ResponseWriter) {
	a.clientError(w, http.StatusNotFound)
}

func (a *app) newTemplateData(r *http.Request) templateData {
	return templateData{
		Version:     a.version[:7],
		CurrentYear: time.Now().Year(),

		// 1. retrieve the value for the flash key
		// 2. and delete the key in one step
		// 3. add flash message to the template data
		Flash: a.sessionManager.PopString(r.Context(), "flash"),

		// add authentication status to the template data
		IsAuthenticated: a.isAuthenticated(r),

		// add CSRF token to the template data
		CSRFToken: nosurf.Token(r),
	}
}

// isAuthenticated checks if the request is from an authenticated user
func (a *app) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		// key not found in ctx, or value was not a boolean
		return false
	}
	return isAuthenticated
}

func (a *app) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	ts, ok := a.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		a.serverError(w, r, err)
		return
	}

	// stage 1: write template into buffer
	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		log.Printf("render err=%v+\n", err)
		a.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)

	// stage 2: write rendered content
	buf.WriteTo(w)
}

func (app *app) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		// no body, or body is too large to process
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		return err
	}

	return nil
}
