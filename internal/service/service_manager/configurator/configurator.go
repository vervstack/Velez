package configurator

import (
	"github.com/docker/docker/client"
	"go.verv.tech/matreshka-be/pkg/matreshka_be_api"
)

type Configurator struct {
	matreshka_be_api.MatreshkaBeAPIClient
	dockerAPI client.CommonAPIClient
}

func New(
	matreshka matreshka_be_api.MatreshkaBeAPIClient,
	docker client.CommonAPIClient,
) (c *Configurator) {
	return &Configurator{
		MatreshkaBeAPIClient: matreshka,
		dockerAPI:            docker,
	}
}
