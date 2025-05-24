package steps

import (
	"context"

	"github.com/docker/docker/api/types"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients"
)

type stepPrepareImage struct {
	docker clients.Docker

	imageName string
	result    *types.ImageInspect
}

func PrepareImageStep(nodeClients clients.NodeClients, imageName string, result *types.ImageInspect) *stepPrepareImage {
	return &stepPrepareImage{
		docker:    nodeClients.Docker(),
		imageName: imageName,
		result:    result,
	}
}

func (s *stepPrepareImage) Do(ctx context.Context) error {
	image, err := s.docker.PullImage(ctx, s.imageName)
	if err != nil {
		return rerrors.Wrap(err, "error pulling image")
	}

	*s.result = image

	return nil
}
