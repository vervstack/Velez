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

// UnscopedEnvironment is the owner recorded for allocations made through the
// non-environment-scoped API (GetPort/LockPort). Every environment draws from
// the same shared pool - environments deliberately do NOT get partitioned port
// ranges, they only get collision-aware ownership tracking.
const (
	UnscopedEnvironment = ""
)

type portManagerImpl struct {
	m         sync.Mutex
	freePorts map[uint32]bool
	// portOwners records which environment currently holds each locked port.
	// A port is in this map only while it's locked; unlocking removes the
	// entry. Ports allocated by internal, node-wide callers are owned by
	// UnscopedEnvironment.
	portOwners map[uint32]string

	holdM       sync.Mutex
	pausedPorts map[uint32]bool
}

func NewPortManager(availablePorts []int, usedPorts []uint32) PortManager {
	pm := &portManagerImpl{
		freePorts:   make(map[uint32]bool, len(availablePorts)),
		portOwners:  make(map[uint32]string, len(availablePorts)),
		pausedPorts: make(map[uint32]bool, len(availablePorts)),
	}

	for _, p := range availablePorts {
		pm.freePorts[uint32(p)] = false
	}

	for _, p := range usedPorts {
		pm.freePorts[p] = true
		pm.portOwners[p] = UnscopedEnvironment
	}

	return pm
}

func (p *portManagerImpl) GetPort() (uint32, error) {
	return p.GetPortForEnvironment(UnscopedEnvironment)
}

// GetPortForEnvironment hands out a free port and records environment as its
// owner. The pool is shared across every environment, so a port already held
// by another environment is never handed out here.
func (p *portManagerImpl) GetPortForEnvironment(environment string) (uint32, error) {
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
		p.portOwners[port] = environment

		return port, nil
	}

	// Second pass: reclaim ports that were allocated before but whose containers
	// have since been removed (e.g. between repeated test runs in -count mode).
	// A port still owned by a *different* environment is skipped: its owner may
	// simply not have bound it yet, and handing it to another environment would
	// reintroduce exactly the cross-environment collision this tracking exists
	// to prevent.
	for port, ok := range p.freePorts {
		if !ok {
			continue
		}

		owner, owned := p.portOwners[port]
		if owned && owner != environment {
			continue
		}

		ln, err := lc.Listen(ctx, "tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if err != nil {
			continue
		}

		_ = ln.Close()
		p.freePorts[port] = true
		p.portOwners[port] = environment

		return port, nil
	}

	return 0, ErrNoPortsAvailable
}

func (p *portManagerImpl) LockPort(ports ...uint32) error {
	return p.LockPortForEnvironment(UnscopedEnvironment, ports...)
}

// LockPortForEnvironment locks ports on behalf of environment. A port already
// locked by anyone - this environment or another - is rejected with
// ErrPortAlreadyLocked; collisions are detected across the whole shared pool,
// never per environment.
func (p *portManagerImpl) LockPortForEnvironment(environment string, ports ...uint32) (err error) {
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

			return err
		}

		if isLocked {
			owner, owned := p.portOwners[port]
			if owned && owner != environment {
				err = errors.Wrapf(ErrPortAlreadyLocked,
					"port %d is held by environment '%s'", port, owner)

				return err
			}

			err = errors.Wrap(ErrPortAlreadyLocked)

			return err
		}

		p.freePorts[port] = true
		p.portOwners[port] = environment
		pL = append(pL, port)
	}

	return nil
}

func (p *portManagerImpl) UnlockPorts(ports []uint32) {
	p.m.Lock()

	for _, item := range ports {
		p.freePorts[item] = false
		delete(p.portOwners, item)
	}

	p.m.Unlock()
}

// PortOwner reports which environment currently holds port, and whether it's
// held at all.
func (p *portManagerImpl) PortOwner(port uint32) (string, bool) {
	p.m.Lock()
	defer p.m.Unlock()

	owner, ok := p.portOwners[port]

	return owner, ok
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
