package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:embed all:dist
var distFS embed.FS

func NewHandler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	return &spaHandler{fs: sub}
}

type spaHandler struct {
	fs fs.FS
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	if path != "" {
		f, err := h.fs.Open(path)
		if err == nil {
			err = f.Close()
			if err != nil {
				log.Err(err).
					Msg("error closing file in client spa handler")
			}
			http.FileServer(http.FS(h.fs)).ServeHTTP(w, r)
			return
		}
	}

	http.ServeFileFS(w, r, h.fs, "index.html")
}
