package v1

import (
	"context"
	"strings"

	errors "github.com/Red-Sock/trace-errors"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

func (c *containerManager) getImage(ctx context.Context, name string) (*velez_api.Image, error) {
	req := domain.ImageListRequest{Name: name}
	if !strings.HasSuffix(name, "latest") {
		// check if specified version already exists
		list, err := listImages(ctx, c.docker, req)
		if err != nil {
			return nil, errors.Wrap(err, "error trying to listImages local images")
		}

		if len(list) != 0 {
			return list[0], nil
		}
	}

	image, err := pullImage(ctx, c.docker, req)
	if err != nil {
		return nil, errors.Wrap(err, "error trying to listImages local images")
	}

	return image, nil
}
