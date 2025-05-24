package steps

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/image"
	"go.redsock.ru/evon"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"go.vervstack.ru/matreshka/pkg/matreshka_be_api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service"
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
	pref := extractPrefix(c.result.Meta.Name)
	if pref == nil {
		switch {
		case c.image.Config.Labels[labels.MatreshkaConfigLabel] == "true":
			pref = toolbox.ToPtr(matreshka_be_api.ConfigTypePrefix_verv)
		case isPostgresByImageTags(c.image.RepoTags):
			pref = toolbox.ToPtr(matreshka_be_api.ConfigTypePrefix_pg)
		default:
			pref = toolbox.ToPtr(matreshka_be_api.ConfigTypePrefix_kv)
		}
	}

	c.result.Meta.ConfType = *pref

	switch *pref {
	case matreshka_be_api.ConfigTypePrefix_verv:
		return c.assembleVervConfig(ctx)
	case matreshka_be_api.ConfigTypePrefix_pg:
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
		Name:    appendPrefix(matreshka_be_api.ConfigTypePrefix_pg, c.req.Name),
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
		Name:    appendPrefix(matreshka_be_api.ConfigTypePrefix_verv, c.result.Meta.Name),
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

func isPostgresByImageTags(tags []string) bool {
	for _, tag := range tags {
		if strings.Contains(tag, "postgres") {
			return true
		}
	}

	return false
}

func extractPrefix(name string) *matreshka_be_api.ConfigTypePrefix {
	for id, pref := range matreshka_be_api.ConfigTypePrefix_name {
		if strings.HasPrefix(name, pref) {
			return toolbox.ToPtr(matreshka_be_api.ConfigTypePrefix(id))
		}
	}

	return nil
}

func appendPrefix(prefix matreshka_be_api.ConfigTypePrefix, name string) string {
	if !strings.HasPrefix(name, prefix.String()) {
		name = prefix.String() + "_" + name
	}

	return name
}
