package main

import (
	"html/template"
	"path/filepath"

	"github.com/tullo/snptx/pkg/models"
)

type templateData struct {
	CurrentYear int
	Snippet     *models.Snippet
	Snippets    []*models.Snippet
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {

	cache := map[string]*template.Template{}

	// slice of filepaths with the extension '.page.tmpl'
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {

		// extract the file name
		name := filepath.Base(page)

		// parse the page template file in to a template set
		ts, err := template.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// add any 'layout' templates to the template set
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		// add any 'partial' templates to the template set
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		// add the template set to the cache
		cache[name] = ts
	}

	return cache, nil
}
