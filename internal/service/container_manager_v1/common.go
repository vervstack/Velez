package container_manager_v1

import (
	"context"
	"strings"

	errors "github.com/Red-Sock/trace-errors"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

func (c *ContainerManager) getImage(ctx context.Context, imageName string) (*velez_api.Image, error) {
	req := domain.ImageListRequest{Name: imageName}

	if !strings.HasSuffix(imageName, "latest") {
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
