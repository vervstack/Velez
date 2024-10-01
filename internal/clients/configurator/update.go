package configurator

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka-be/pkg/matreshka_be_api"
)

func (c *Configurator) UpdateConfig(ctx context.Context, serviceName string, config matreshka.AppConfig) (err error) {

	postConfig := &matreshka_be_api.PostConfig_Request{
		ServiceName: serviceName,
	}

	postConfig.Content, err = config.Marshal()
	if err != nil {
		return errors.Wrap(err, "error marshalling config")
	}
	_, err = c.matreshkaClient.PostConfig(ctx, postConfig)
	if err != nil {
		return errors.Wrap(err, "error patching config")
	}

	return nil
}
