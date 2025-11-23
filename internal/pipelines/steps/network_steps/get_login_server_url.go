package network_steps

import (
	"context"

	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

type getLoginServerUrlStep struct {
	responsePtr *string
}

func GetLoginServerUrl(responsePtr *string) steps.Step {
	return &getLoginServerUrlStep{
		responsePtr: responsePtr,
	}
}

func (g *getLoginServerUrlStep) Do(_ context.Context) error {
	// TODO For multiple nodes implement different urls
	*g.responsePtr = "http://headscale.verv:8080"
	return nil
}
