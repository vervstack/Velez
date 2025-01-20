package deploy_steps

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	dockerContainerStatusRunning = "running"
)

type healthcheckStep struct {
	docker clients.Docker

	req *velez_api.CreateSmerd_Request

	state *domain.LaunchSmerdState
}

func HealthcheckStep(
	docker clients.Docker,
	req *velez_api.CreateSmerd_Request,
	dp *domain.LaunchSmerdState,
) *healthcheckStep {
	return &healthcheckStep{
		docker: docker,

		req:   req,
		state: dp,
	}
}

func (h *healthcheckStep) Do(ctx context.Context) error {
	if h.req.Healthcheck == nil {
		return nil
	}

	if h.state.ContainerId == nil {
		return rerrors.New("container was not created")
	}

	errC := make(chan error)
	go func() {
		defer close(errC)

		for i := uint32(0); i < h.req.Healthcheck.Retries; i++ {
			time.Sleep(time.Duration(h.req.Healthcheck.IntervalSecond) * time.Second)

			cont, err := h.docker.ContainerInspect(ctx, *h.state.ContainerId)
			if err != nil {
				errC <- err
				return
			}
			if cont.State.Health == nil {
				continue
			}

			if cont.State.Status == dockerContainerStatusRunning {
				errC <- nil
				return
			}
		}
	}()

	err := <-errC
	if err != nil {
		return rerrors.Wrap(err, "error during healthcheck")
	}

	return nil
}
