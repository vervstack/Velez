package config_manager

import (
	"context"
	"os"
	"path"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"

	"github.com/godverv/Velez/internal/config"
)

func (c *Configurator) UpdateConfig(ctx context.Context, serviceName string, config *matreshka.AppConfig) error {
	raw, err := config.Marshal()
	if err != nil {
		return errors.Wrap(err, "error marshalling config")
	}

	err = c.updateConfigInApi(ctx, serviceName, raw)
	if err != nil {
		return errors.Wrap(err, "error updating config by api")
	}

	err = c.updateContainerConfig(serviceName, raw)
	if err != nil {
		return errors.Wrap(err, "error updating config in smerd")
	}

	return nil
}

func (c *Configurator) updateConfigInApi(ctx context.Context, serviceName string, raw []byte) error {
	_, err := c.matreshkaClient.PatchConfigRaw(ctx,
		&matreshka_api.PatchConfigRaw_Request{
			Raw:         raw,
			ServiceName: serviceName,
		})
	if err != nil {
		return errors.Wrap(err, "error patching config")
	}

	return nil
}

func (c *Configurator) updateContainerConfig(serviceName string, raw []byte) error {
	mp := path.Join(c.getMountPoint(serviceName), path.Base(config.ProdConfigPath))

	err := os.WriteFile(mp, raw, 777)
	if err != nil {
		return errors.Wrap(err, "error writing config to mount point")
	}

	return nil
}
