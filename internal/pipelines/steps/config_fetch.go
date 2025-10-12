package steps

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

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/utils/configutils"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type fetchConfigStep struct {
	configService service.ConfigurationService

	req   *domain.LaunchSmerd
	image *image.InspectResponse

	result *domain.ConfigMount
}

func FetchConfig(
	services service.Services,

	req *domain.LaunchSmerd,
	image *image.InspectResponse,
	result *domain.ConfigMount,
) *fetchConfigStep {
	return &fetchConfigStep{
		configService: services.ConfigurationService(),

		req:   req,
		image: image,

		result: result,
	}
}

func (c *fetchConfigStep) Do(ctx context.Context) (err error) {
	err = c.validate()
	if err != nil {
		return rerrors.Wrap(err, "error during validation")
	}

	if c.req.Config == nil {
		c.req.Config = &velez_api.MatreshkaConfigSpec{
			ConfigName: toolbox.ToPtr(c.req.Name),
		}
	}

	*c.result, err = c.do(ctx, c.req.Config)
	if err != nil {
		return rerrors.Wrap(err, "error getting config to mount")
	}

	return nil
}

func (c *fetchConfigStep) do(ctx context.Context, spec *velez_api.MatreshkaConfigSpec) (mount domain.ConfigMount, err error) {
	mount.FilePath = spec.SystemPath

	fillMeta(c.image, &mount)

	configName := toolbox.Coalesce(spec.ConfigName, &c.req.Name)
	mount.Meta.Name = configutils.AppendPrefix(mount.Meta.ConfType, *configName)

	mount.Meta.Version = spec.ConfigVersion

	mount.Meta.Format = *toolbox.Coalesce(spec.ConfigFormat, &mount.Meta.Format)

	if c.req.IgnoreConfig {
		return mount, nil
	}

	if mount.Meta.Format == velez_api.ConfigFormat_env {
		err = c.setEnv(ctx, mount.Meta)
	} else {
		mount.Content, err = c.getPlain(ctx, mount.Meta)
	}

	if err != nil {
		return mount, rerrors.Wrap(err, "error getting config from matreshka")
	}

	return mount, nil
}

func (c *fetchConfigStep) setEnv(ctx context.Context, meta domain.ConfigMeta) error {
	envEvon, err := c.configService.GetEnvFromApi(ctx, meta)
	if err != nil {
		code := status.Code(err)
		if code != codes.NotFound {
			return rerrors.Wrap(err, "error getting matreshka config from matreshka api")
		}
		err = nil
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
		err = nil
	}

	return plainConfig, nil
}

func (c *fetchConfigStep) validate() error {
	if c.result == nil {
		return rerrors.New("empty result pointer")
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
