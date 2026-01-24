package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticFS embed.FS

// StaticFileHandler returns an HTTP handler for serving embedded static files.
func StaticFileHandler() http.Handler {
	// Get the static subdirectory from the embedded filesystem
	subFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		// This should never happen with embedded files
		panic("failed to access static subdirectory: " + err.Error())
	}

	return http.FileServer(http.FS(subFS))
}
