package deploy_manager

import (
	"context"
	"time"

	errors "github.com/Red-Sock/trace-errors"

	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	dockerContainerStatusRunning = "running"
)

func (d *DeployManager) Healthcheck(ctx context.Context, contId string, hc *velez_api.Container_Healthcheck) error {
	errC := make(chan error)

	go func() {
		defer close(errC)

		for i := uint32(0); i < hc.Retries; i++ {
			time.Sleep(time.Duration(hc.IntervalSecond) * time.Second)

			cont, err := d.docker.ContainerInspect(ctx, contId)
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
		return errors.Wrap(err, "error during healthcheck")
	}

	return nil
}
