package steps

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/image"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"go.vervstack.ru/matreshka/pkg/matreshka_be_api"

	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type prepareVervConfig struct {
	docker        clients.Docker
	configService service.ConfigurationService
	portManager   clients.PortManager

	req   *domain.LaunchSmerd
	image *image.InspectResponse

	lockedPorts []uint32
}

func PrepareVervConfig(
	nodeClients clients.NodeClients,
	srv service.Services,

	req *domain.LaunchSmerd,
	image *image.InspectResponse,
) *prepareVervConfig {
	return &prepareVervConfig{
		docker:        nodeClients.Docker(),
		configService: srv.ConfigurationService(),
		portManager:   nodeClients.PortManager(),

		req:   req,
		image: image,
	}
}

func (p *prepareVervConfig) Do(ctx context.Context) error {
	if p.image.Config.Labels == nil {
		p.image.Config.Labels = make(map[string]string)
	}

	err := p.enrichWithMatreshkaConfig(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error enriching with verv data")
	}

	if p.req.UseImagePorts {
		err = p.getPortsFromImage()
		if err != nil {
			return rerrors.Wrap(err, "error locking ports")
		}
	}

	err = p.lockPorts()
	if err != nil {
		return rerrors.Wrap(err, "error locking ports")
	}

	p.req.Env[matreshka.VervName] = p.req.GetName()
	p.req.Labels[labels.CreatedWithVelezLabel] = "true"

	for _, networks := range p.req.Settings.Network {
		err = dockerutils.CreateNetwork(ctx, p.docker, networks.NetworkName)
		if err != nil {
			return rerrors.Wrap(err, "error creating network: %s", networks.NetworkName)
		}
	}

	return nil
}

func (p *prepareVervConfig) Rollback(_ context.Context) error {
	for _, port := range p.lockedPorts {
		if !p.portManager.UnHoldPort(port) {
			p.portManager.UnlockPorts(p.lockedPorts)
		}
	}
	return nil
}

func (p *prepareVervConfig) enrichWithMatreshkaConfig(ctx context.Context) error {
	if p.req.IgnoreConfig {
		p.req.Labels[labels.MatreshkaConfigLabel] = "false"
		return nil
	}

	cfgMeta := domain.ConfigMeta{
		Name:    p.req.Name,
		Version: p.req.ConfigVersion,
	}

	// TODO think about it
	t := matreshka_be_api.ConfigTypePrefix_kv

	if p.image.Config.Labels[labels.MatreshkaConfigLabel] == "true" {
		t = matreshka_be_api.ConfigTypePrefix_verv
	}

	cfgMeta.Name = t.String() + "_" + cfgMeta.Name

	envVars, err := p.configService.GetEnvFromApi(ctx, cfgMeta)
	if err != nil {
		return rerrors.Wrap(err, "error getting matreshka config from matreshka api")
	}

	for _, e := range envVars.InnerNodes {
		if len(e.InnerNodes) == 0 && e.Value != nil {
			_, exists := p.req.Env[e.Name]
			if !exists {
				p.req.Env[e.Name] = fmt.Sprint(e.Value)
			}
		}
	}

	return nil
}

func (p *prepareVervConfig) getPortsFromImage() error {
	portsInReq := map[uint32]*velez_api.Port{}
	for _, port := range p.req.Settings.Ports {
		portsInReq[port.ServicePortNumber] = port
	}

	for port := range p.image.Config.ExposedPorts {
		portVal := uint32(port.Int())
		_, ok := portsInReq[portVal]
		if ok {
			continue
		}

		portBind := &velez_api.Port{
			ServicePortNumber: portVal,
			Protocol:          velez_api.Port_Protocol(velez_api.Port_Protocol_value[port.Proto()]),
		}
		p.req.Settings.Ports = append(p.req.Settings.Ports, portBind)
		portsInReq[portVal] = portBind
	}

	return nil
}

func (p *prepareVervConfig) lockPorts() (err error) {
	p.lockedPorts = make([]uint32, 0, len(p.req.Settings.Ports))

	for _, imagePort := range p.req.Settings.Ports {
		if imagePort.ExposedTo == nil {
			var port uint32
			port, err = p.portManager.GetPort()
			imagePort.ExposedTo = &port
		} else {
			ok := p.portManager.UnHoldPort(*imagePort.ExposedTo)
			if !ok {
				err = p.portManager.LockPort(*imagePort.ExposedTo)
			}
		}
		if err != nil {
			err = rerrors.Wrap(err, "error locking host port")
			return
		}

		p.lockedPorts = append(p.lockedPorts, *imagePort.ExposedTo)
	}

	return nil
}
