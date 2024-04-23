package config_manager

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"
)

func (c *Configurator) UpdateConfig(ctx context.Context, serviceName string, config matreshka.AppConfig) error {
	raw, err := config.Marshal()
	if err != nil {
		return errors.Wrap(err, "error marshalling config")
	}

	_, err = c.matreshkaClient.PatchConfigRaw(
		ctx,
		&matreshka_api.PatchConfigRaw_Request{
			Raw:         raw,
			ServiceName: serviceName,
		})
	if err != nil {
		return errors.Wrap(err, "error patching config")
	}

	return nil
}
