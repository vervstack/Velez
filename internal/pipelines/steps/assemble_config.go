package steps

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/service"
)

type assembleConfigStep struct {
	docker        clients.Docker
	configService service.ConfigurationService
	contId        *string
	serviceName   string

	result *matreshka.AppConfig
}

func AssembleConfigStep(
	docker clients.Docker,
	configService service.ConfigurationService,
	contId *string,
	serviceName string,

	result *matreshka.AppConfig,
) *assembleConfigStep {
	return &assembleConfigStep{
		docker:        docker,
		configService: configService,
		contId:        contId,
		serviceName:   serviceName,

		result: result,
	}
}

func (c *assembleConfigStep) Do(ctx context.Context) error {
	if c.contId == nil {
		return rerrors.New("empty container id")
	}

	configFromContainer, err := c.configService.GetFromContainer(ctx, *c.contId)
	if err != nil {
		return rerrors.Wrap(err, "error getting matreshka config from container")
	}

	configFromApi, err := c.configService.GetFromApi(ctx, c.serviceName)
	if err != nil {
		code := status.Code(err)
		if code != codes.NotFound {
			return rerrors.Wrap(err, "error getting matreshka config from matreshka api")
		}

		configFromApi = matreshka.NewEmptyConfig()
	}

	*c.result = matreshka.MergeConfigs(configFromApi, configFromContainer)

	return nil
}
