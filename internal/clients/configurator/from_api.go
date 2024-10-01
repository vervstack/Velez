package configurator

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka-be/pkg/matreshka_be_api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Configurator) GetFromApi(ctx context.Context, serviceName string) (matreshka.AppConfig, error) {
	var apiConfig matreshka.AppConfig

	req := &matreshka_be_api.GetConfig_Request{
		ServiceName: serviceName,
	}
	matreshkaConfig, err := c.matreshkaClient.GetConfig(ctx, req)
	if err != nil {
		e, ok := status.FromError(err)
		if ok && e.Code() == codes.NotFound {
			return matreshka.AppConfig{}, nil
		}

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
