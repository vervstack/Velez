package configurator

import (
	"context"

	errors "go.redsock.ru/rerrors"
	"go.verv.tech/matreshka"
	"go.verv.tech/matreshka-be/pkg/matreshka_be_api"
)

func (c *Configurator) GetFromApi(ctx context.Context, serviceName string) (matreshka.AppConfig, error) {
	var apiConfig matreshka.AppConfig

	req := &matreshka_be_api.GetConfig_Request{
		ServiceName: serviceName,
	}
	matreshkaConfig, err := c.MatreshkaBeAPIClient.GetConfig(ctx, req)
	if err != nil {
		return matreshka.AppConfig{}, errors.Wrap(err, "error obtaining raw config")
	}

	if matreshkaConfig.Config == nil {
		return matreshka.NewEmptyConfig(), nil
	}

	err = apiConfig.Unmarshal(matreshkaConfig.Config)
	if err != nil {
		return matreshka.AppConfig{}, errors.Wrap(err, "error unmarshalling config from api")
	}

	return apiConfig, nil
}
