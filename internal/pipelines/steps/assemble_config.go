package steps

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/image"
	"go.redsock.ru/evon"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"go.vervstack.ru/matreshka/pkg/matreshka_be_api"
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
	image *image.InspectResponse

	contId *string

	result *domain.AppConfig
}

func AssembleConfigStep(
	nodeClients clients.NodeClients,
	services service.Services,
	contId *string,
	req domain.LaunchSmerd,
	image *image.InspectResponse,
	result *domain.AppConfig,
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

	if c.image.Config == nil {
		return nil
	}
	switch {
	case c.image.Config.Labels[labels.MatreshkaConfigLabel] == "true":
		return c.assembleVervConfig(ctx)
	case isPostgres(c.image.RepoTags):
		return c.assemblePostgresConfig(ctx)
	default:
		return c.assembleKvConfig(ctx)
	}

}
func (c *assembleConfigStep) assembleKvConfig(_ context.Context) error {
	c.result.Meta.ConfType = matreshka_be_api.ConfigTypePrefix_kv

	return nil
}

func (c *assembleConfigStep) assemblePostgresConfig(ctx context.Context) (err error) {
	c.result.Meta.ConfType = matreshka_be_api.ConfigTypePrefix_pg

	cfgMeta := domain.ConfigMeta{
		Name:    matreshka_be_api.ConfigTypePrefix_pg.String() + "_" + c.req.Name,
		Version: c.req.ConfigVersion,
	}

	c.result.Content, err = c.configService.GetEnvFromApi(ctx, cfgMeta)
	if err != nil {
		code := status.Code(err)
		if code != codes.NotFound {
			return rerrors.Wrap(err, "error getting matreshka config from matreshka api")
		}
	}

	return nil
}

func (c *assembleConfigStep) assembleVervConfig(ctx context.Context) error {
	c.result.Meta.ConfType = matreshka_be_api.ConfigTypePrefix_verv

	configFromContainer, err := c.configService.GetFromContainer(ctx, *c.contId)
	if err != nil {
		return rerrors.Wrap(err, "error getting matreshka config from container")
	}

	cfgMeta := domain.ConfigMeta{
		Name:    matreshka_be_api.ConfigTypePrefix_verv.String() + "_" + c.req.Name,
		Version: c.req.ConfigVersion,
	}

	configFromApi, err := c.configService.GetFromApi(ctx, cfgMeta)
	if err != nil {
		code := status.Code(err)
		if code != codes.NotFound {
			return rerrors.Wrap(err, "error getting matreshka config from matreshka api")
		}

		configFromApi = matreshka.NewEmptyConfig()
		err = nil
	}

	vervCfg := matreshka.MergeConfigs(configFromApi, configFromContainer)
	marshalled, err := evon.MarshalEnv(&vervCfg)
	if err != nil {
		return rerrors.Wrap(err, "error marshalling matreshka config")
	}

	c.result.Content = marshalled
	return nil
}

func isPostgres(tags []string) bool {
	for _, tag := range tags {
		if strings.Contains(tag, "postgres") {
			return true
		}
	}

	return false
}
