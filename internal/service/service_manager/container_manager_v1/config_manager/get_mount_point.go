package config_manager

import (
	"os"
	"path"

	errors "github.com/Red-Sock/trace-errors"
)

func (c *Configurator) GetMountPoint(serviceName string) (string, error) {
	configFolder := c.getMountPoint(serviceName)

	err := os.MkdirAll(configFolder, 0777)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return "", errors.Wrap(err, "error creating folder for configs")
		}
	}

	return configFolder, nil
}

func (c *Configurator) getMountPoint(serviceName string) string {
	return path.Join(c.vervVolumePath, serviceName, FolderName)
}
