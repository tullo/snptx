//go:build go1.16
// +build go1.16

package main

import (
	"io/fs"
	"testing"

	"github.com/pkg/errors"
)

func TestEmbedHTMLTemplates(t *testing.T) {
	html := &Embed{
		FS:  webUI,
		Dir: "ui/html",
	}
	uifs, err := html.Sub()
	if err != nil {
		t.Error(err)
	}
	files, err := fs.ReadDir(uifs, ".")
	if err != nil {
		t.Error(errors.Wrap(err, "reading directory from embed fs"))
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		b, err := fs.ReadFile(uifs, f.Name())
		if err != nil {
			t.Error(errors.Wrapf(err, "reading embeded file %s", f.Name()))
		}
		t.Logf("%s, (%v) bytes", f.Name(), len(b))
	}
}
