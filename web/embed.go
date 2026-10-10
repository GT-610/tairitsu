// Package web exposes the built frontend bundle to the Go binary.
//
// This file lives beside the frontend sources because go:embed can only reach
// files inside its own package directory. The bundle is produced by the
// frontend build and compiled into the server, so a release ships as one
// process with no separate web server in front of it.
package web

import (
	"embed"
	"io/fs"
)

// distDir is the directory produced by the frontend build.
const distDir = "dist"

//go:embed all:dist
var bundled embed.FS

// Assets returns the embedded frontend bundle rooted at the build output.
func Assets() (fs.FS, error) {
	return fs.Sub(bundled, distDir)
}

// IsBuilt reports whether a real frontend build was embedded. It is false when
// the binary was compiled without running the frontend build first.
func IsBuilt() bool {
	_, err := fs.Stat(bundled, distDir+"/index.html")
	return err == nil
}
