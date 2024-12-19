package smerd_launcher

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	errors "go.redsock.ru/rerrors"
	"go.verv.tech/matreshka"
	"google.golang.org/grpc/codes"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	CreatedWithVelezLabel = "CREATED_WITH_VELEZ"

	MatreshkaConfigLabel = "MATRESHKA_CONFIG_ENABLED"
)

type SmerdLauncher struct {
	docker        clients.Docker
	deployManager clients.DeployManager
	portManager   clients.PortManager
	configService service.ConfigurationService
}

func New(
	nodeClients clients.NodeClients,
	configService service.ConfigurationService,
) *SmerdLauncher {
	return &SmerdLauncher{
		docker:        nodeClients.Docker(),
		deployManager: nodeClients.DeployManager(),
		portManager:   nodeClients.PortManager(),

		configService: configService,
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

	if image.Config.Labels[MatreshkaConfigLabel] == "true" {
		err = c.enrichWithMatreshkaConfig(ctx, req)
		if err != nil {
			return "", errors.Wrap(err, "error enriching with verv data")
		}
	} else {
		// TODO заиспользовать ручку из VERV-75 для получения конфигурации ресурса
	}

	if req.UseImagePorts {
		err = c.getPortsFromImage(image, req)
		if err != nil {
			return "", errors.Wrap(err, "error locking ports")
		}
	}

	lockedPorts, err := c.lockPorts(req)
	if err != nil {
		err = errors.Wrap(err, "error locking ports", codes.ResourceExhausted)
		return
	}

	defer func() {
		if err != nil {
			c.portManager.UnlockPorts(lockedPorts)
		}
	}()

	req.Env[matreshka.VervName] = req.GetName()
	req.Labels[CreatedWithVelezLabel] = "true"

	cont, err = c.deployManager.Create(ctx, req)
	if err != nil {
		err = errors.Wrap(err, "error creating container")
		return
	}

	err = c.deployManager.Healthcheck(ctx, cont.ID, req.Healthcheck)
	if err != nil {
		err = errors.Wrap(err, "error performing healthcheck")
		return
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
	if req.IgnoreConfig {
		req.Labels[MatreshkaConfigLabel] = "false"
		return nil
	}

	envs, err := c.configService.GetEnvFromApi(ctx, req.GetName())
	if err != nil {
		return errors.Wrap(err, "error getting matreshka config from matreshka api")
	}
	serviceName := strings.ToUpper(req.GetName())
	for _, e := range envs {
		if len(e.InnerNodes) == 0 && e.Value != nil {
			req.Env[serviceName+"_"+e.Name] = fmt.Sprint(e.Value)
		}
	}

	return nil
}

func (c *SmerdLauncher) getPortsFromImage(
	image types.ImageInspect,
	req *velez_api.CreateSmerd_Request) error {

	portsInReq := map[uint32]*velez_api.Port{}
	for _, port := range req.Settings.Ports {
		portsInReq[port.ServicePortNumber] = port
	}

	for port := range image.Config.ExposedPorts {
		portVal := uint32(port.Int())
		_, ok := portsInReq[portVal]
		if ok {
			continue
		}

		portBind := &velez_api.Port{
			ServicePortNumber: portVal,
			Protocol:          velez_api.Port_Protocol(velez_api.Port_Protocol_value[port.Proto()]),
		}
		req.Settings.Ports = append(req.Settings.Ports, portBind)
		portsInReq[portVal] = portBind
	}

	return nil
}

func (c *SmerdLauncher) lockPorts(req *velez_api.CreateSmerd_Request) (lockedPorts []uint32, err error) {
	lockedPorts = make([]uint32, 0, len(req.Settings.Ports))

	defer func() {
		if err != nil {
			c.portManager.UnlockPorts(lockedPorts)
		}
	}()

	for _, p := range req.Settings.Ports {
		if p.ExposedTo == nil {
			var port uint32
			port, err = c.portManager.GetPort()
			p.ExposedTo = &port
		} else {
			err = c.portManager.LockPort(*p.ExposedTo)
		}
		if err != nil {
			err = errors.Wrap(err, "error locking host port")
			return
		}

		lockedPorts = append(lockedPorts, *p.ExposedTo)
	}

	return lockedPorts, nil
}
