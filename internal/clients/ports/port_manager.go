package ports

import (
	"context"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	errors "go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/config"
)

var (
	ErrUnavailablePort   = errors.New("port is not available for velez")
	ErrPortAlreadyLocked = errors.New("port is already obtained")
	ErrNoPortsAvailable  = errors.New("no ports available")
)

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

func (p *PortManager) LockPort(ports ...uint32) (err error) {
	if len(ports) == 0 {
		return nil
	}
	pL := make([]uint32, 0, len(ports))

	p.m.Lock()
	defer func() {
		if err != nil {
			p.UnlockPorts(pL)
		}
	}()
	defer p.m.Unlock()

	for _, port := range ports {
		isLocked, ok := p.ports[port]
		if !ok {
			err = errors.Wrap(ErrUnavailablePort)
			return
		}
		if isLocked {
			err = errors.Wrap(ErrPortAlreadyLocked)
			return
		}

		if len(pL) == cap(pL) {
			break
		}
	}

	if len(pL) != cap(pL) {
		err = errors.Wrap(ErrNoPortsAvailable)
		return
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
