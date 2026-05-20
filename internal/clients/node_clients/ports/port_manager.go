package ports

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/docker/docker/api/types/container"
	errors "go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/config"
)

var (
	ErrUnavailablePort   = errors.New("port is not available for velez")
	ErrPortAlreadyLocked = errors.New("port is already obtained")
	ErrNoPortsAvailable  = errors.New("no ports available")
)

type portManagerImpl struct {
	m         sync.Mutex
	freePorts map[uint32]bool

	holdM       sync.Mutex
	pausedPorts map[uint32]bool
}

func NewPortManager(ctx context.Context, cfg config.Config, docker *docker.Docker) (PortManager, error) {
	pm := &portManagerImpl{
		freePorts:   make(map[uint32]bool, len(cfg.Environment.AvailablePorts)),
		pausedPorts: make(map[uint32]bool, len(cfg.Environment.AvailablePorts)),
	}

	containerList, err := docker.Client().ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error listing container")
	}

	for _, item := range cfg.Environment.AvailablePorts {
		pm.freePorts[uint32(item)] = false
	}

	for _, item := range containerList {
		for _, port := range item.Ports {
			if port.PublicPort != 0 {
				pm.freePorts[uint32(port.PublicPort)] = true
			}
		}
	}

	return pm, nil
}

func (p *portManagerImpl) GetPort() (uint32, error) {
	p.m.Lock()
	defer p.m.Unlock()

	// First pass: only consider ports not already marked as in-use in memory.
	// This avoids the TOCTOU race between allocation and Docker binding.
	for port, ok := range p.freePorts {
		if ok {
			continue
		}
		ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if err != nil {
			p.freePorts[port] = true
			continue
		}
		ln.Close()
		p.freePorts[port] = true
		return port, nil
	}

	// Second pass: reclaim ports that were allocated before but whose containers
	// have since been removed (e.g. between repeated test runs in -count mode).
	for port, ok := range p.freePorts {
		if !ok {
			continue
		}
		ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if err != nil {
			continue
		}
		ln.Close()
		p.freePorts[port] = true
		return port, nil
	}

	return 0, ErrNoPortsAvailable
}

func (p *portManagerImpl) LockPort(ports ...uint32) (err error) {
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
		isLocked, ok := p.freePorts[port]
		if !ok {
			err = errors.Wrap(ErrUnavailablePort)
			return
		}
		if isLocked {
			err = errors.Wrap(ErrPortAlreadyLocked)
			return
		}

		p.freePorts[port] = true
		pL = append(pL, port)
	}

	return nil
}

func (p *portManagerImpl) UnlockPorts(ports []uint32) {
	p.m.Lock()

	for _, item := range ports {
		p.freePorts[item] = false
	}

	p.m.Unlock()
}

func (p *portManagerImpl) HoldPort(port uint32) bool {
	p.holdM.Lock()
	wasOnHold := p.pausedPorts[port]
	p.pausedPorts[port] = true
	p.holdM.Unlock()

	return !wasOnHold
}

func (p *portManagerImpl) UnHoldPort(port uint32) bool {
	p.holdM.Lock()
	wasOnHold := p.pausedPorts[port]
	p.pausedPorts[port] = false
	p.holdM.Unlock()

	return wasOnHold
}
