// +build go1.16

package main

import (
	"embed"
	"io/fs"

	"github.com/pkg/errors"
)

// Embed defines a struct used to embed files natively in Go 1.16.
type Embed struct {
	FS  embed.FS
	Dir string
}

// Sub returns a FS corresponding to the subtree rooted at <Embed.Dir>.
func (e *Embed) Sub() (fs.FS, error) {
	var efs fs.FS = e.FS
	if e.Dir != "" {
		var err error
		efs, err = fs.Sub(efs, e.Dir)
		if err != nil {
			return nil, errors.Wrapf(err, "opening subdirectory in embed fs: %s", e.Dir)
		}
	}

	return efs, nil
}
