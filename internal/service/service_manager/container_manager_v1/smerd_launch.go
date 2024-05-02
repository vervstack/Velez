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
	c.normalizeCreateRequest(req)

	image, err := dockerutils.PullImage(ctx, c.docker, req.ImageName, false)
	if err != nil {
		return "", errors.Wrap(err, "error pulling image")
	}

	// TODO https://redsock.youtrack.cloud/issue/VERV-56
	if req.Settings != nil {
		err = c.portManager.LockPorts(req.Settings.Ports)
		if err != nil {
			return "", errors.Wrap(err, "error getting ports on host side")
		}
	}

	var cont *types.Container

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

	//if cfg.Healthcheck != nil {
	//	hcC := c.waitHealthcheck(ctx, cont.ID, *cfg.Healthcheck)
	//	err = <-hcC
	//	if err != nil {
	//		return "", errors.Wrap(err, "healthcheck failed")
	//	}
	//}

	return cont.ID, nil
}

func (c *ContainerManager) normalizeCreateRequest(req *velez_api.CreateSmerd_Request) {
	if req.Settings == nil {
		req.Settings = &velez_api.Container_Settings{}
	}

	if req.Hardware == nil {
		req.Hardware = &velez_api.Container_Hardware{}
	}
}

func (c *ContainerManager) waitHealthcheck(
	ctx context.Context,
	containerId string,
	hc container.HealthConfig,
) chan error {
	errC := make(chan error)

	go func() {
		defer close(errC)

		for i := 0; i < hc.Retries; i++ {
			time.Sleep(hc.Interval)

			cont, err := c.docker.ContainerInspect(ctx, containerId)
			if err != nil {
				errC <- err
				return
			}
			if cont.State.Health == nil {
				continue
			}

			if cont.State.Health.Status == "healthy" {
				errC <- nil
				return
			}
		}
	}()

	return errC
}
