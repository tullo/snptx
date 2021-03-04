// +build go1.16

package main

import (
	"html/template"
	"io/fs"
	"path"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/tullo/snptx/internal/forms"
	"github.com/tullo/snptx/internal/snippet"
	"github.com/tullo/snptx/internal/user"
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

func newTemplateCache(dir string) (map[string]*template.Template, error) {

	cache := map[string]*template.Template{}

	html := &Embed{
		FS:  webUI,
		Dir: dir,
	}

	uifs, err := html.Sub()
	if err != nil {
		return nil, errors.Wrap(err, "creating embed file system")
	}

	files, err := fs.ReadDir(uifs, ".")
	if err != nil {
		return nil, errors.Wrap(err, "reading directory from embed fs")
	}

	var pages []string

	for i := range files {
		if files[i].IsDir() {
			continue
		}
		if ok, _ := path.Match("*.page.tmpl", files[i].Name()); ok {
			pages = append(pages, files[i].Name())
		}
	}

	for _, page := range pages {

		// extract the file name
		name := filepath.Base(page)

		// parse the page template file in to a template set
		ts, err := template.New(name).Funcs(functions).ParseFS(uifs, page)
		if err != nil {
			return nil, err
		}

		// add any 'layout' templates to the template set
		ts, err = ts.ParseFS(uifs, "*.layout.tmpl")
		if err != nil {
			return nil, err
		}

		// add any 'partial' templates to the template set
		ts, err = ts.ParseFS(uifs, "*.partial.tmpl")
		if err != nil {
			return nil, err
		}

		// add the template set to the cache
		cache[name] = ts
	}

	return cache, nil
}
