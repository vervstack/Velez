package container_manager_v1

import (
	"context"
	"strings"

	"github.com/docker/docker/client"
	matreshkaPb "github.com/godverv/matreshka-be/pkg/matreshka_api"
	"github.com/pkg/errors"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/internal/service/container_manager_v1/port_manager"
	"github.com/godverv/Velez/pkg/velez_api"
)

type containerManager struct {
	docker          client.CommonAPIClient
	matreshkaClient matreshkaPb.MatreshkaBeAPIClient
	portManager     *port_manager.PortManager
}

func NewContainerManager(cfg config.Config, docker client.CommonAPIClient, apiClient matreshkaPb.MatreshkaBeAPIClient) (*containerManager, error) {
	c := &containerManager{
		docker:          docker,
		matreshkaClient: apiClient,
	}

	var err error

	c.portManager, err = port_manager.NewPortManager(context.Background(), cfg, docker)
	if err != nil {
		return nil, errors.Wrap(err, "error getting port manager")
	}

	return c, nil
}

func (c *containerManager) getImage(ctx context.Context, name string) (*velez_api.Image, error) {
	req := domain.ImageListRequest{Name: name}
	if !strings.HasSuffix(name, "latest") {
		// check if specified version already exists
		list, err := dockerutils.ListImages(ctx, c.docker, req)
		if err != nil {
			return nil, errors.Wrap(err, "error trying to ListImages local images")
		}

		if len(list) != 0 {
			return list[0], nil
		}
	}

	image, err := dockerutils.PullImage(ctx, c.docker, req)
	if err != nil {
		return nil, errors.Wrap(err, "error trying to ListImages local images")
	}

	return image, nil
}
