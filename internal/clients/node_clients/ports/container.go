package ports

import (
	"sync/atomic"

	"go.redsock.ru/rerrors"
)

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
	port, err := (*c.impl.Load()).GetPort()
	if err != nil {
		return 0, rerrors.Wrap(err, "error getting port")
	}

	return port, nil
}

func (c *Container) LockPort(p ...uint32) error {
	err := (*c.impl.Load()).LockPort(p...)
	if err != nil {
		return rerrors.Wrap(err, "error locking port")
	}

	return nil
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
