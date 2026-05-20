package ports

type PortManager interface {
	GetPort() (uint32, error)
	LockPort(ports ...uint32) error
	UnlockPorts(ports []uint32)

	// HoldPort returns true if port was not on hold (you set it on hold)
	// returns false if port is already on hold
	HoldPort(port uint32) bool
	// UnHoldPort return true if port was on hold (you take it)
	// returns false if port is not on hold
	UnHoldPort(port uint32) bool
}
