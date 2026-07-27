package config_steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/image"
	"go.redsock.ru/evon"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/utils/configutils"
)

type fetchConfigStep struct {
	configService service.ConfigurationService

	req   *domain.LaunchSmerd
	image *image.InspectResponse

	mounts *[]domain.FileMountPoint
}

func FetchConfig(
	services service.Services,

	req *domain.LaunchSmerd,
	image *image.InspectResponse,
	mounts *[]domain.FileMountPoint,
) *fetchConfigStep {
	return &fetchConfigStep{
		configService: services.ConfigurationService(),

		req:   req,
		image: image,

		mounts: mounts,
	}
}

func (c *fetchConfigStep) Do(ctx context.Context) (err error) {
	err = c.validate()
	if err != nil {
		return rerrors.Wrap(err, "error during validation")
	}

	if c.req.Config == nil {
		c.req.Config = &velez_api.CreateSmerd_Request_Verv{
			Verv: &velez_api.MatreshkaConfigSpec{
				ConfigName: toolbox.ToPtr(c.req.Name),
			},
		}
	}

	vervCfg := c.req.GetVerv()
	if vervCfg != nil {
		mount, err := c.doVerv(ctx, vervCfg)
		if err != nil {
			return rerrors.Wrap(err, "error getting config to mount")
		}

		*c.mounts = append(*c.mounts, mount)

		return nil
	}

	plainCfg := c.req.GetPlain()
	if plainCfg != nil {
		return c.doPlain(plainCfg)
	}

	return rerrors.New("only verv config supported for now")
}

func (c *fetchConfigStep) doVerv(
	ctx context.Context,
	spec *velez_api.MatreshkaConfigSpec,
) (domain.FileMountPoint, error) {
	var cfgMount domain.ConfigMount

	cfgMount.FilePath = spec.SystemPath

	fillMeta(c.image, &cfgMount)

	configName := toolbox.Coalesce(spec.ConfigName, &c.req.Name)

	cfgMount.Meta.Name = configutils.AppendPrefix(cfgMount.Meta.ConfType, *configName)

	cfgMount.Meta.Version = spec.ConfigVersion

	cfgMount.Meta.Format = *toolbox.Coalesce(spec.ConfigFormat, &cfgMount.Meta.Format)

	if c.req.IgnoreConfig {
		return cfgMount.FileMountPoint, nil
	}

	var err error

	if cfgMount.Meta.Format == velez_api.ConfigFormat_env {
		err = c.setEnv(ctx, cfgMount.Meta)
	} else {
		cfgMount.Content, err = c.getPlain(ctx, cfgMount.Meta)
	}

	if err != nil {
		return domain.FileMountPoint{}, rerrors.Wrap(err, "error getting config from matreshka")
	}

	return cfgMount.FileMountPoint, nil
}

func (c *fetchConfigStep) doPlain(spec *velez_api.PlainConfigSpec) error {
	for sysPath, content := range spec.GetConfigs() {
		p := sysPath

		*c.mounts = append(*c.mounts, domain.FileMountPoint{
			FilePath: &p,
			Content:  content,
		})
	}

	return nil
}

func (c *fetchConfigStep) setEnv(ctx context.Context, meta domain.ConfigMeta) error {
	envEvon, err := c.configService.GetEnvFromApi(ctx, meta)
	if err != nil {
		code := status.Code(err)
		if code != codes.NotFound {
			return rerrors.Wrap(err, "error getting matreshka config from matreshka api")
		}
	}

	ns := evon.NodeStorage{}
	ns.AddNode(envEvon)

	for _, n := range ns {
		if n.Value == nil {
			continue
		}

		c.req.Env[n.Name] = fmt.Sprint(n.Value)
	}

	return nil
}

func (c *fetchConfigStep) getPlain(ctx context.Context, meta domain.ConfigMeta) ([]byte, error) {
	plainConfig, err := c.configService.GetPlainFromApi(ctx, meta)
	if err != nil {
		code := status.Code(err)
		if code != codes.NotFound {
			return nil, rerrors.Wrap(err, "error getting matreshka config from matreshka api")
		}
	}

	return plainConfig, nil
}

func (c *fetchConfigStep) validate() error {
	if c.mounts == nil {
		return rerrors.New("empty mounts pointer")
	}

	if c.image.Config == nil {
		return rerrors.New("empty image config")
	}

	return nil
}

func isVervImage(image *image.InspectResponse) bool {
	return image.Config.Labels[labels.MatreshkaConfigLabel] == "true"
}

func isPostgresByImageTags(tags []string) bool {
	for _, tag := range tags {
		if strings.Contains(tag, "postgres") {
			return true
		}
	}

	return false
}
