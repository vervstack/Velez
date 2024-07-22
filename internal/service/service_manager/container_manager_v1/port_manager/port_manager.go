package port_manager

import (
	"context"
	"sync"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/pkg/velez_api"
)

var ErrNoPortsAvailable = errors.New("no ports available")

type PortManager struct {
	m sync.Mutex

	ports map[uint32]bool
}

func NewPortManager(ctx context.Context, cfg config.Config, docker client.CommonAPIClient) (*PortManager, error) {
	ports, err := config.GetAvailablePorts(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error reading available ports from config")
	}

	pm := &PortManager{
		ports: make(map[uint32]bool, len(ports)),
	}

	containerList, err := docker.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error listing container")
	}

	for _, item := range ports {
		pm.ports[item] = false
	}

	for _, item := range containerList {
		for _, port := range item.Ports {
			if port.PublicPort != 0 {
				pm.ports[uint32(port.PublicPort)] = true
			}
		}
	}

	return pm, nil
}

func (p *PortManager) GetPort() (uint32, error) {
	p.m.Lock()
	defer p.m.Unlock()

	for port, ok := range p.ports {
		if ok {
			continue
		}

		portCopy := port
		p.ports[portCopy] = true
		return portCopy, nil
	}

	return 0, ErrNoPortsAvailable
}

func (p *PortManager) LockPorts(ports []*velez_api.PortBindings) error {
	if len(ports) == 0 {
		return nil
	}

	pL := make([]uint32, 0, len(ports))
	for range ports {
		port, err := p.GetPort()
		if err != nil {
			return errors.Wrap(err)
		}

		pL = append(pL, port)
		if len(pL) == cap(pL) {
			break
		}
	}

	if len(pL) != cap(pL) {
		p.UnlockPorts(pL)
		return ErrNoPortsAvailable
	}

	for idx, portBind := range ports {
		portBind.Host = uint32(pL[idx])
		portBind.Protoc = velez_api.PortBindings_tcp
	}

	return nil
}

func (p *PortManager) UnlockPorts(ports []uint32) {
	p.m.Lock()

	for _, item := range ports {
		p.ports[item] = false
	}

	p.m.Unlock()
}

func (p *PortManager) UnlockFromSettings(ports []*velez_api.PortBindings) {
	p.m.Lock()

	for _, item := range ports {
		p.ports[item.Host] = false
	}

	p.m.Unlock()
}
