package main

import (
	"github.com/rs/zerolog/log"

	"go.vervstack.ru/Velez/internal/app"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	err = a.Start()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
