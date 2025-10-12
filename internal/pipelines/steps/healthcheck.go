package steps

import (
	"context"
	"time"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/domain"
)

const (
	dockerContainerStatusRunning = "running"
)

type healthcheckStep struct {
	dockerAPI client.APIClient

	req         *domain.LaunchSmerd
	containerId *string
}

func Healthcheck(
	nodeClients clients.NodeClients,
	req *domain.LaunchSmerd,
	containerId *string,
) *healthcheckStep {
	return &healthcheckStep{
		dockerAPI: nodeClients.Docker().Client(),

		req:         req,
		containerId: containerId,
	}
}

func (h *healthcheckStep) Do(ctx context.Context) error {
	if h.req.Healthcheck == nil {
		return nil
	}

	if h.containerId == nil && *h.containerId == "" {
		return rerrors.New("container was not created")
	}

	errC := make(chan error)
	go func() {
		defer close(errC)

		for i := uint32(0); i < h.req.Healthcheck.Retries; i++ {
			time.Sleep(time.Duration(h.req.Healthcheck.IntervalSecond) * time.Second)

			cont, err := h.dockerAPI.ContainerInspect(ctx, *h.containerId)
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
