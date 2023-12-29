package container_manager_v1

import (
	"context"
	"strings"

	"github.com/docker/docker/client"
	"github.com/pkg/errors"

	"github.com/godverv/Velez/internal/client/docker/dockerutils"
	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

type containerManager struct {
	docker client.CommonAPIClient
}

func NewContainerManager(docker client.CommonAPIClient) (*containerManager, error) {
	return &containerManager{
		docker: docker,
	}, nil
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
