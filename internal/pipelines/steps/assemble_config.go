package steps

import (
	"context"

	"github.com/docker/docker/api/types"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/internal/domain/labels"
	"github.com/godverv/Velez/internal/service"
)

type assembleConfigStep struct {
	docker        clients.Docker
	configService service.ConfigurationService

	req   domain.LaunchSmerd
	image *types.ImageInspect

	contId *string

	result *matreshka.AppConfig
}

func AssembleConfigStep(
	nodeClients clients.NodeClients,
	services service.Services,
	contId *string,
	req domain.LaunchSmerd,
	image *types.ImageInspect,
	result *matreshka.AppConfig,
) *assembleConfigStep {
	return &assembleConfigStep{
		docker:        nodeClients.Docker(),
		configService: services.ConfigurationService(),

		req:   req,
		image: image,

		contId: contId,
		result: result,
	}
}

func (c *assembleConfigStep) Do(ctx context.Context) error {
	if c.contId == nil {
		return rerrors.New("empty container id")
	}

	if c.result == nil {
		return rerrors.New("empty result pointer")
	}

	if c.image.Config == nil || c.image.Config.Labels[labels.MatreshkaConfigLabel] != "true" {
		return nil
	}

	configFromContainer, err := c.configService.GetFromContainer(ctx, *c.contId)
	if err != nil {
		return rerrors.Wrap(err, "error getting matreshka config from container")
	}

	configFromApi, err := c.configService.GetFromApi(ctx, c.req.Name)
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
