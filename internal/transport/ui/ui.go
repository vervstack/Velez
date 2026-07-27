package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:embed all:dist
var frontend embed.FS

func NewServer() http.Handler {
	mux := http.NewServeMux()

	stripped, err := fs.Sub(frontend, "dist")
	if err != nil {
		log.Fatal().Err(err).Msg("error reading embedded frontend")
	}

	ffs := http.FileServer(http.FS(stripped))
	mux.Handle("/", spaFallback(stripped, ffs))

	return mux
}

// spaFallback rewrites requests for paths that don't exist in the embedded
// build (e.g. client-side routes like /cp) to "/", so the file server returns
// index.html and the browser router can take over instead of a hard 404.
func spaFallback(fsys fs.FS, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if requested == "" {
			requested = "."
		}

		_, err := fs.Stat(fsys, requested)
		if err != nil {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}

		next.ServeHTTP(w, r)
	})
}
