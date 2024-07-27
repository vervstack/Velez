package deploy_manager

import (
	"github.com/docker/docker/client"
)

type DeployManager struct {
	docker client.CommonAPIClient
}

func New(docker client.CommonAPIClient) *DeployManager {
	return &DeployManager{
		docker: docker,
	}
}
