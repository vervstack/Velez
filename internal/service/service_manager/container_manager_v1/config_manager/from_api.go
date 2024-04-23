package config_manager

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"
)

func (c *Configurator) GetFromApi(ctx context.Context, serviceName string) (matreshka.AppConfig, error) {
	var apiConfig matreshka.AppConfig

	matreshkaConfig, err := c.matreshkaClient.GetConfigRaw(ctx,
		&matreshka_api.GetConfigRaw_Request{
			ServiceName: serviceName,
		},
	)
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
