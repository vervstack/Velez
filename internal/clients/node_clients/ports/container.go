package ports

import "sync/atomic"

type Container struct {
	impl atomic.Pointer[PortManager]
}

func NewContainer(initial PortManager) *Container {
	c := &Container{}
	c.Set(initial)
	return c
}

func (c *Container) Set(impl PortManager) {
	c.impl.Store(&impl)
}

func (c *Container) GetPort() (uint32, error) {
	return (*c.impl.Load()).GetPort()
}

func (c *Container) LockPort(p ...uint32) error {
	return (*c.impl.Load()).LockPort(p...)
}

func (c *Container) UnlockPorts(p []uint32) {
	(*c.impl.Load()).UnlockPorts(p)
}

func (c *Container) HoldPort(p uint32) bool {
	return (*c.impl.Load()).HoldPort(p)
}

func (c *Container) UnHoldPort(p uint32) bool {
	return (*c.impl.Load()).UnHoldPort(p)
}
