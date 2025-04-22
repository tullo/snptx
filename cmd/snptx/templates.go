//go:build go1.16
// +build go1.16

package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/tullo/snptx/internal/forms"
	"github.com/tullo/snptx/internal/snippet"
	"github.com/tullo/snptx/internal/user"
	"github.com/tullo/snptx/ui"
)

type templateData struct {
	CSRFToken       string
	CurrentYear     int
	Flash           string
	Form            *forms.Form
	IsAuthenticated bool
	Snippet         *snippet.Info
	Snippets        []snippet.Info
	User            *user.Info
	Version         string
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	// Convert the time to UTC before formatting it.
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

func shortID(s string) string {
	if len(s) < 8 {
		return s
	}
	return s[:8]
}

var functions = template.FuncMap{
	"humanDate": humanDate,
	"shortID":   shortID,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {

		// extract the file name
		name := filepath.Base(page)

		// parse the page template file in to a template set
		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		// add the template set to the cache
		cache[name] = ts
	}

	return cache, nil
}
