// Package web embeds the built frontend (frontend/dist) into the server
// binary so that one artifact serves both the API and the UI (ADR-0008).
//
// `make build` copies frontend/dist into ./dist before compiling. In a
// plain `go build` without that step the directory holds only .gitkeep and
// Dist reports false, which the HTTP layer turns into a 503.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embedded embed.FS

// Dist returns the built frontend as a filesystem rooted at index.html,
// and whether a build is actually present.
func Dist() (fs.FS, bool) {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		return nil, false
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil, false
	}
	return sub, true
}
