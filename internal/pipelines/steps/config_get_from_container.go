package steps

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/image"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type getConfigFromContainerStep struct {
	docker        clients.Docker
	configService service.ConfigurationService

	req *domain.LaunchSmerd

	contId *string
	image  *image.InspectResponse
	result *domain.ConfigMount
}

func GetConfigFromContainerStep(
	nodeClients clients.NodeClients,
	services service.Services,

	req *domain.LaunchSmerd,

	contId *string,
	image *image.InspectResponse,
	result *domain.ConfigMount,
) *getConfigFromContainerStep {
	return &getConfigFromContainerStep{
		docker:        nodeClients.Docker(),
		configService: services.ConfigurationService(),

		req: req,

		image:  image,
		contId: contId,
		result: result,
	}
}

func (c *getConfigFromContainerStep) Do(ctx context.Context) error {
	err := c.validate()
	if err != nil {
		return rerrors.Wrap(err, "error during validation")
	}

	c.result.Meta.Name = toolbox.Coalesce(c.result.Meta.Name, c.req.Name)

	fillMeta(c.image, c.result)

	spec := c.getSpec(c.result)

	if spec == nil {
		return rerrors.New("can't guess specification. Set it up manually via `MatreshkaConfigSpec config` field")
	}

	c.result.Content, err = c.extractConfig(ctx, spec)
	if err != nil {
		return rerrors.Wrap(err, "error getting config to mount")
	}

	return nil
}

func (c *getConfigFromContainerStep) extractConfig(ctx context.Context,
	spec *velez_api.MatreshkaConfigSpec) (res []byte, err error) {
	if spec.SystemPath == nil {
		return nil, rerrors.New("no target path for config specified")
	}

	res, err = dockerutils.ReadFromContainer(ctx, c.docker, *c.contId, *spec.SystemPath)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return res, nil
}

func (c *getConfigFromContainerStep) getSpec(mount *domain.ConfigMount) *velez_api.MatreshkaConfigSpec {
	if c.req.Config != nil {
		return c.req.Config
	}

	spec := &velez_api.MatreshkaConfigSpec{
		ConfigName:    &mount.Meta.Name,
		ConfigVersion: mount.Meta.Version,
		ConfigFormat:  &mount.Meta.Format,
		SystemPath:    mount.FilePath,
	}

	return spec
}

func (c *getConfigFromContainerStep) validate() error {
	if c.contId == nil {
		return rerrors.New("empty container id")
	}

	if c.result == nil {
		return rerrors.New("empty result pointer")
	}

	if c.image.Config == nil {
		return rerrors.New("empty image config")
	}
	return nil
}

func fillMeta(img *image.InspectResponse, mount *domain.ConfigMount) {
	switch {
	case isVervImage(img):
		mount.Meta.Format = matreshka_api.Format_yaml
		mount.Meta.ConfType = matreshka_api.ConfigTypePrefix_verv
		mount.FilePath = toolbox.Coalesce(
			mount.FilePath,
			toolbox.ToPtr("/app/config/config.yaml"),
		)
	case isPostgresByImageTags(img.RepoTags):
		mount.Meta.ConfType = matreshka_api.ConfigTypePrefix_pg
		mount.Meta.Format = matreshka_api.Format_env

		// TODO add support for postgres customization
		//mount.FilePath = toolbox.Coalesce(
		//	mount.FilePath,
		//	toolbox.ToPtr("/var/lib/postgresql/data/postgresql.conf"),
		//)
	default:
		mount.Meta.ConfType = matreshka_api.ConfigTypePrefix_plain
	}

	if strings.HasSuffix(toolbox.FromPtr(mount.FilePath), matreshka_api.Format_yaml.String()) {
		mount.Meta.Format = matreshka_api.Format_yaml
	} else {
		mount.Meta.Format = matreshka_api.Format_env
	}

}
