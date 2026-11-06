package main

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"

	"go.vervstack.ru/Velez/internal/app"
	"go.vervstack.ru/Velez/internal/selfswap"
)

func main() {
	if os.Getenv(selfswap.EnvSelfSwap) == "1" {
		err := selfswap.Run(context.Background())
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		return
	}

	a, err := app.New()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	err = a.Start()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
