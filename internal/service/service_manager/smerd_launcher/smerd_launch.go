package smerd_launcher

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/godverv/matreshka"

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
	configManager clients.Configurator
}

func New(cl clients.Clients) *SmerdLauncher {
	return &SmerdLauncher{
		docker:        cl.Docker(),
		deployManager: cl.DeployManager(),
		portManager:   cl.PortManager(),
		configManager: cl.Configurator(),
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
		err = c.enrichWithMatreshkaConfig(ctx, req)
		if err != nil {
			return "", errors.Wrap(err, "error enriching with verv data")
		}
	} else {
		// TODO заиспользовать ручку из VERV-75 для получения конфигурации ресурса
	}

	req.Env[matreshka.VervName] = req.GetName()
	req.Labels[CreatedWithVelezLabel] = "true"

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

	if req.Labels == nil {
		req.Labels = make(map[string]string)
	}

	return nil
}

func (c *SmerdLauncher) enrichWithMatreshkaConfig(ctx context.Context, req *velez_api.CreateSmerd_Request) error {
	matreshkaConfig, err := c.configManager.GetFromApi(ctx, req.GetName())
	if err != nil {
		return errors.Wrap(err, "error getting matreshka config from matreshka api")
	}

	for _, srv := range matreshkaConfig.Servers {
		req.Settings.Ports = append(req.Settings.Ports,
			&velez_api.PortBindings{
				Container: uint32(srv.GetPort()),
				Protoc:    velez_api.PortBindings_tcp,
			})
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
