package v1

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/domain"
)

type ContainerManager struct {
	dockerClient client.CommonAPIClient

	ports map[uint16]bool
}

func NewContainerManager(cfg config.Config, dockerClient client.CommonAPIClient) (*ContainerManager, error) {
	cm := &ContainerManager{
		dockerClient: dockerClient,
	}

	var err error
	cm.ports, err = cm.portsList(cfg, dockerClient)
	if err != nil {
		return nil, errors.Wrap(err, "error obtaining ports lists")
	}

	return cm, nil
}

func (c *ContainerManager) Up(ctx context.Context, container domain.CreateContainer) (domain.Container, error) {
	return domain.Container{}, nil
}

func (c *ContainerManager) portsList(cfg config.Config, dockerApi client.CommonAPIClient) (map[uint16]bool, error) {
	ap, err := config.AvailablePorts(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error obtaining available ports")
	}

	containerList, err := dockerApi.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, errors.Wrap(err, "error listing containers")
	}

	m := make(map[uint16]bool)
	for _, c := range containerList {
		for _, p := range c.Ports {
			if p.PublicPort == 0 {
				continue
			}

			m[p.PublicPort] = true
		}

	}

	for _, a := range ap {
		m[a] = false
	}

	return m, nil
}
