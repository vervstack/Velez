package container_manager_v1

import (
	"github.com/docker/docker/client"
)

type containerManager struct {
	docker client.CommonAPIClient
}

func NewContainerManager(docker client.CommonAPIClient) (*containerManager, error) {
	return &containerManager{
		docker: docker,
	}, nil
}
