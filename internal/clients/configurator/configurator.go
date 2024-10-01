package configurator

import (
	"github.com/docker/docker/client"
	"github.com/godverv/matreshka-be/pkg/matreshka_be_api"
)

type Configurator struct {
	matreshkaClient matreshka_be_api.MatreshkaBeAPIClient
	dockerAPI       client.CommonAPIClient
}

func New(
	matreshka matreshka_be_api.MatreshkaBeAPIClient,
	docker client.CommonAPIClient,
) (c *Configurator) {
	return &Configurator{
		matreshkaClient: matreshka,
		dockerAPI:       docker,
	}
}
