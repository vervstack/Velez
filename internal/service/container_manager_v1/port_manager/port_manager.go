package port_manager

import (
	"context"
	"sync"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/godverv/Velez/internal/config"
)

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

	containerList, err := docker.ContainerList(ctx, types.ContainerListOptions{})
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
