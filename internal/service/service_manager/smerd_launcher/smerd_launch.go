package smerd_launcher

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	CreatedWithVelezLabel = "CREATED_WITH_VELEZ"

	matreshkaConfigLabel = "MATRESHKA_CONFIG_ENABLED"
)

type SmerdLauncher struct {
	docker        clients.Docker
	deployManager clients.DeployManager
	portManager   clients.PortManager
}

func New(cl clients.Clients) *SmerdLauncher {
	return &SmerdLauncher{
		docker:        cl.Docker(),
		deployManager: cl.DeployManager(),
		portManager:   cl.PortManager(),
	}
}

func (c *SmerdLauncher) LaunchSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (id string, err error) {
	err = c.normalizeCreateRequest(req)
	if err != nil {
		return "", errors.Wrap(err, "error normalizing create request")
	}

	image, err := c.docker.PullImage(ctx, req.ImageName)
	if err != nil {
		return "", errors.Wrap(err, "error pulling image")
	}

	var cont *types.ContainerJSON

	if image.Labels[matreshkaConfigLabel] == "true" {
		// TODO Do create verv here
	}

	cont, err = c.deployManager.Create(ctx, req)
	if err != nil {
		return "", errors.Wrap(err, "error creating container")
	}

	err = c.deployManager.Healthcheck(ctx, cont.ID, req.Healthcheck)
	if err != nil {
		return "", errors.Wrap(err, "error performing healthcheck")
	}

	return cont.ID, nil
}

func (c *SmerdLauncher) normalizeCreateRequest(req *velez_api.CreateSmerd_Request) error {
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

	req.Labels[CreatedWithVelezLabel] = "true"

	return nil
}

//func ()  {
//for _, p := range req.Settings.Ports {
//	if p.Host == 0 {
//		var err error
//		p.Host, err = c.portManager.GetPort()
//		if err != nil {
//			return errors.Wrap(err, "error getting host port")
//		}
//	} else {
//		err := c.portManager.LockPorts(req.Settings.Ports)
//		if err != nil {
//			return errors.Wrap(err, "error locking ports for container")
//		}
//	}
//
//}

//req.Labels[CreatedWithVelezLabel] = "true"
//}
