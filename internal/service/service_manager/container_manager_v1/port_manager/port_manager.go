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

	ports map[uint16]bool
}

func NewPortManager(ctx context.Context, cfg config.Config, docker client.CommonAPIClient) (*PortManager, error) {
	ports, err := config.GetAvailablePorts(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error reading available ports from config")
	}

	pm := &PortManager{
		ports: make(map[uint16]bool, len(ports)),
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
				pm.ports[port.PublicPort] = true
			}
		}
	}

	return pm, nil
}

func (p *PortManager) GetPort() *uint16 {
	p.m.Lock()
	defer p.m.Unlock()

	for port, ok := range p.ports {
		if ok {
			continue
		}

		port := port
		return &port
	}

	return nil
}

func (p *PortManager) FillPorts(ports []*velez_api.PortBindings) error {
	if len(ports) == 0 {
		return nil
	}

	p.m.Lock()
	defer p.m.Unlock()

	pL := make([]uint16, 0, len(ports))

	for port, ok := range p.ports {
		if ok {
			continue
		}

		pL = append(pL, port)
		if len(pL) == cap(pL) {
			break
		}
	}

	if len(pL) != cap(pL) {
		return ErrNoPortsAvailable
	}

	for i := range pL {
		p.ports[pL[i]] = true
		ports[i].Host = uint32(pL[i])
	}

	return nil
}

func (p *PortManager) Free(ports []uint16) {
	p.m.Lock()

	for _, item := range ports {
		p.ports[item] = false
	}

	p.m.Unlock()
}

func (p *PortManager) FreeFromSettings(ports []*velez_api.PortBindings) {
	p.m.Lock()

	for _, item := range ports {
		p.ports[uint16(item.Host)] = false
	}

	p.m.Unlock()
}
