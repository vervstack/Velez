package ports

type PortManager interface {
	// GetPort is GetPortForEnvironment(UnscopedEnvironment).
	GetPort() (uint32, error)
	// GetPortForEnvironment hands out a free port from the single shared pool
	// and records environment as its owner, so the same port can never be
	// handed to two environments at once.
	GetPortForEnvironment(environment string) (uint32, error)

	// LockPort is LockPortForEnvironment(UnscopedEnvironment, ports...).
	LockPort(ports ...uint32) error
	// LockPortForEnvironment locks ports on behalf of environment. A port
	// already held by a different environment fails with ErrPortAlreadyLocked
	// naming the current owner.
	LockPortForEnvironment(environment string, ports ...uint32) error

	UnlockPorts(ports []uint32)

	// PortOwner returns the environment currently holding port and whether
	// it's held at all.
	PortOwner(port uint32) (string, bool)

	// HoldPort returns true if port was not on hold (you set it on hold)
	// returns false if port is already on hold
	HoldPort(port uint32) bool
	// UnHoldPort return true if port was on hold (you take it)
	// returns false if port is not on hold
	UnHoldPort(port uint32) bool
}
