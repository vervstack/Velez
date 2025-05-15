package configurator

import (
	"context"

	errors "go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"go.vervstack.ru/matreshka/pkg/matreshka_be_api"

	"github.com/godverv/Velez/internal/domain"
)

func (c *Configurator) GetFromApi(ctx context.Context, meta domain.ConfigMeta) (matreshka.AppConfig, error) {
	var apiConfig matreshka.AppConfig

	req := &matreshka_be_api.GetConfig_Request{
		ConfigName: meta.ServiceName,
		Version:    meta.CfgVersion,
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
