package container_manager

import (
	"context"
	stderrs "errors"

	errors "go.redsock.ru/rerrors"
	"go.verv.tech/matreshka"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/godverv/Velez/internal/service/service_manager/smerd_launcher/shared"
	"github.com/godverv/Velez/internal/service/service_manager/smerd_launcher/steps"
	"github.com/godverv/Velez/pkg/velez_api"
)

const configFetchingPostfix = "_config_scanning"

func (c *ContainerManager) FetchConfig(ctx context.Context, req *velez_api.FetchConfig_Request) (*matreshka.AppConfig, error) {
	_, err := c.docker.PullImage(ctx, req.ImageName)
	if err != nil {
		return nil, errors.Wrap(err, "error pulling image")
	}

	createReq := &velez_api.CreateSmerd_Request{
		Name:      req.ServiceName + configFetchingPostfix,
		ImageName: req.ImageName,
		Settings:  &velez_api.Container_Settings{},
	}
	dp := &shared.DeployProcess{}
	err = steps.LaunchContainer(c.docker, createReq, dp).Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error creating container")
	}
	defer func() {
		_, dropErr := c.DropSmerds(ctx, &velez_api.DropSmerd_Request{
			Uuids: []string{dp.Container.ID},
		})
		if dropErr != nil {
			err = stderrs.Join(err, errors.Wrap(dropErr, "error dropping config scanning smerd"))
		}
	}()

	configFromContainer, err := c.configService.GetFromContainer(ctx, dp.Container.ID)
	if err != nil {
		return nil, errors.Wrap(err, "error getting matreshka config from container")
	}

	configFromApi, err := c.configService.GetFromApi(ctx, req.GetServiceName())
	if err != nil {
		code := status.Code(err)
		if code != codes.NotFound {
			return nil, errors.Wrap(err, "error getting matreshka config from matreshka api")
		}

		configFromApi = matreshka.NewEmptyConfig()
	}

	matreshkaConfig := matreshka.MergeConfigs(configFromApi, configFromContainer)

	err = c.configService.UpdateConfig(ctx, req.ServiceName, matreshkaConfig)
	if err != nil {
		return nil, errors.Wrap(err, "error updating config")
	}

	return &matreshkaConfig, nil
}
