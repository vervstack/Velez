package deploy_steps

import (
	"context"

	"go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/domain"
)

type stepPrepareImage struct {
	docker clients.Docker

	imageName string
	resp      *domain.LaunchSmerdState
}

func PrepareImageStep(docker clients.Docker, imageName string, resp *domain.LaunchSmerdState) *stepPrepareImage {
	return &stepPrepareImage{
		docker:    docker,
		imageName: imageName,
		resp:      resp,
	}
}

func (s *stepPrepareImage) Do(ctx context.Context) (err error) {
	s.resp.Image, err = s.docker.PullImage(ctx, s.imageName)
	if err != nil {
		return rerrors.Wrap(err, "error pulling image")
	}

	return nil
}
