package container_manager_v1

import (
	errors "github.com/Red-Sock/trace-errors"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (c *ContainerManager) lockPorts(settings *velez_api.Container_Settings) error {
	if settings == nil {
		return nil
	}

	err := c.portManager.FillPorts(settings.Ports)
	if err != nil {
		return errors.Wrap(err, "error getting ports on host side")
	}

	return nil
}

func (c *ContainerManager) freePorts(settings *velez_api.Container_Settings) {
	if settings == nil {
		return
	}

	c.portManager.FreeFromSettings(settings.Ports)
}
