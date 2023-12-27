package v1

import (
	"context"
	"strings"

	errors "github.com/Red-Sock/trace-errors"

	"github.com/godverv/Velez/internal/domain"
)

func (c *containerManager) getImage(ctx context.Context, name string) (domain.Image, error) {
	if !strings.HasSuffix(name, "latest") {
		// check if specified version already exists
		list, err := listImages(ctx, c.docker, domain.ImageListRequest{ImageName: name})
		if err != nil {
			return domain.Image{}, errors.Wrap(err, "error trying to listImages local images")
		}

		if len(list) != 0 {
			return list[0], nil
		}
	}

	image, err := pullImage(ctx, c.docker, domain.Image{Name: name})
	if err != nil {
		return domain.Image{}, errors.Wrap(err, "error trying to listImages local images")
	}

	return image, nil
}
