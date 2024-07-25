package container_manager_v1

import (
	"context"
	"time"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	matreshkaConfigLabel = "MATRESHKA_CONFIG_ENABLED"
)

func (c *ContainerManager) LaunchSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (id string, err error) {
	err = c.normalizeCreateRequest(req)
	if err != nil {
		return "", errors.Wrap(err, "error normalizing create request")
	}

	image, err := dockerutils.PullImage(ctx, c.docker, req.ImageName, false)
	if err != nil {
		return "", errors.Wrap(err, "error pulling image")
	}

	var cont *types.ContainerJSON

	if image.Labels[matreshkaConfigLabel] == "true" {
		cont, err = c.containerLauncher.createVerv(ctx, req)
	} else {
		cont, err = c.containerLauncher.createSimple(ctx, req)
	}
	if err != nil {
		return "", errors.Wrap(err, "error creating container")
	}

	err = c.docker.ContainerStart(ctx, cont.ID, container.StartOptions{})
	if err != nil {
		return "", errors.Wrap(err, "error starting container")
	}

	if req.Healthcheck != nil {
		err = c.doHealthcheck(ctx, cont.ID, req.Healthcheck)
		if err != nil {
			return "", errors.Wrap(err, "error during healthcheck")
		}
	}

	return cont.ID, nil
}

func (c *ContainerManager) normalizeCreateRequest(req *velez_api.CreateSmerd_Request) error {
	if req.Settings == nil {
		req.Settings = &velez_api.Container_Settings{}
	}

	if req.Hardware == nil {
		req.Hardware = &velez_api.Container_Hardware{}
	}

	if req.Env == nil {
		req.Env = make(map[string]string)
	}

	for _, p := range req.Settings.Ports {
		if p.Host == 0 {
			var err error
			p.Host, err = c.portManager.GetPort()
			if err != nil {
				return errors.Wrap(err, "error getting host port")
			}
		} else {
			err := c.portManager.LockPorts(req.Settings.Ports)
			if err != nil {
				return errors.Wrap(err, "error locking ports for container")
			}
		}

	}

	if req.Labels == nil {
		req.Labels = make(map[string]string)
	}

	return nil
}

func (c *ContainerManager) doHealthcheck(
	ctx context.Context,
	containerId string,
	hc *velez_api.Container_Healthcheck,
) error {
	errC := make(chan error)

	go func() {
		defer close(errC)

		for i := uint32(0); i < hc.Retries; i++ {
			time.Sleep(time.Duration(hc.IntervalSecond) * time.Second)

			cont, err := c.docker.ContainerInspect(ctx, containerId)
			if err != nil {
				errC <- err
				return
			}
			if cont.State.Health == nil {
				continue
			}

			if cont.State.Status == "running" {
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
