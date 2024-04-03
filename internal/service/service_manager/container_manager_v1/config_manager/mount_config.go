package config_manager

import (
	"context"
	"os"
	"path"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"

	"github.com/godverv/Velez/internal/config"
)

func (c *Configurator) Mount(ctx context.Context, serviceName string) (*matreshka.AppConfig, error) {
	cfg, err := c.getConfig(ctx, serviceName)
	if err != nil {
		return nil, errors.Wrap(err, "error getting config for service")
	}

	b, err := cfg.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling config for service")
	}
	configFolder := path.Join(c.getMountPoint(serviceName), path.Base(config.ProdConfigPath))

	err = os.WriteFile(configFolder, b, 0777)
	if err != nil {
		return nil, errors.Wrap(err, "no folder for ")
	}

	return cfg, nil
}
