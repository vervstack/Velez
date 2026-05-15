package ui

import (
	"embed"
	"io/fs"
	"net/http"

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
	mux.Handle("/", ffs)

	return mux
}
