package steps

import (
	"context"

	"go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/service/service_manager/smerd_launcher/shared"
	"github.com/godverv/Velez/pkg/velez_api"
)

type stepPrepareImage struct {
	docker clients.Docker

	req  *velez_api.CreateSmerd_Request
	resp *shared.DeployProcess
}

func PrepareImageStep(docker clients.Docker, req *velez_api.CreateSmerd_Request, resp *shared.DeployProcess) shared.Step {
	return &stepPrepareImage{
		docker: docker,
		req:    req,
		resp:   resp,
	}
}

func (s *stepPrepareImage) Do(ctx context.Context) (err error) {
	s.resp.Image, err = s.docker.PullImage(ctx, s.req.ImageName)
	if err != nil {
		return rerrors.Wrap(err, "error pulling image")
	}

	return nil
}
