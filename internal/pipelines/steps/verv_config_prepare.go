package steps

import (
	"context"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type prepareVervConfig struct {
	dockerAPI client.APIClient

	configService service.ConfigurationService
	portManager   node_clients.PortManager

	req   *domain.LaunchSmerd
	image *image.InspectResponse

	lockedPorts []uint32
}

func PrepareVervConfig(
	nodeClients node_clients.NodeClients,
	srv service.Services,

	req *domain.LaunchSmerd,
	image *image.InspectResponse,
) *prepareVervConfig {
	return &prepareVervConfig{
		dockerAPI:     nodeClients.Docker().Client(),
		configService: srv.ConfigurationService(),
		portManager:   nodeClients.PortManager(),

		req:   req,
		image: image,
	}
}

func (p *prepareVervConfig) Do(ctx context.Context) (err error) {
	if p.image.Config.Labels == nil {
		p.image.Config.Labels = make(map[string]string)
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

	if !p.req.IgnoreConfig {
		p.req.Env[matreshka.VervName] = p.req.GetName()
	} else {
		if p.image.Config.Labels[labels.MatreshkaConfigLabel] == "true" {
			p.image.Config.Labels[labels.MatreshkaConfigLabel] = "false"
		}
	}

	for name, val := range p.image.Config.Labels {
		p.req.Labels[name] = val
	}

	p.req.Labels[labels.CreatedWithVelezLabel] = "true"

	// Todo: Think about where to get group name
	p.req.Labels[labels.ComposeGroupLabel] = p.req.GetName()
	if p.req.AutoUpgrade {
		p.req.Labels[labels.AutoUpgrade] = "true"
	}

	for _, networks := range p.req.Settings.Network {
		err = dockerutils.CreateNetwork(ctx, p.dockerAPI, networks.NetworkName)
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

func (p *prepareVervConfig) getPortsFromImage() error {
	portsFromReq := map[uint32]*velez_api.Port{}
	for _, port := range p.req.Settings.Ports {
		portsFromReq[port.ServicePortNumber] = port
	}

	for portProtoc := range p.image.Config.ExposedPorts {
		// TODO for some reasone there was a panic

		pp := strings.Split(portProtoc, "/")
		portVal, err := strconv.ParseUint(pp[0], 10, 32)
		if err != nil {
			return rerrors.Wrap(err, "error parsing port")
		}

		_, ok := portsFromReq[uint32(portVal)]
		if ok {
			continue
		}

		portBind := &velez_api.Port{
			ServicePortNumber: uint32(portVal),
			Protocol:          velez_api.Port_Protocol(velez_api.Port_Protocol_value[pp[1]]),
		}

		p.req.Settings.Ports = append(p.req.Settings.Ports, portBind)
		portsFromReq[uint32(portVal)] = portBind
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
