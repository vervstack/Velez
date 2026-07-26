package ports

import (
	"context"
	"fmt"
	"net"
	"sync"

	errors "go.redsock.ru/rerrors"
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

func NewPortManager(availablePorts []int, usedPorts []uint32) PortManager {
	pm := &portManagerImpl{
		freePorts:   make(map[uint32]bool, len(availablePorts)),
		pausedPorts: make(map[uint32]bool, len(availablePorts)),
	}

	for _, p := range availablePorts {
		pm.freePorts[uint32(p)] = false
	}

	for _, p := range usedPorts {
		pm.freePorts[p] = true
	}

	return pm
}

func (p *portManagerImpl) GetPort() (uint32, error) {
	p.m.Lock()
	defer p.m.Unlock()

	lc := net.ListenConfig{}
	ctx := context.Background()

	// First pass: only consider ports not already marked as in-use in memory.
	// This avoids the TOCTOU race between allocation and Docker binding.
	for port, ok := range p.freePorts {
		if ok {
			continue
		}

		ln, err := lc.Listen(ctx, "tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if err != nil {
			p.freePorts[port] = true

			continue
		}

		_ = ln.Close()
		p.freePorts[port] = true

		return port, nil
	}

	// Second pass: reclaim ports that were allocated before but whose containers
	// have since been removed (e.g. between repeated test runs in -count mode).
	for port, ok := range p.freePorts {
		if !ok {
			continue
		}

		ln, err := lc.Listen(ctx, "tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if err != nil {
			continue
		}

		_ = ln.Close()
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
