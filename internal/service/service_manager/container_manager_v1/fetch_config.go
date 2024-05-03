package container_manager_v1

import (
	"context"
	stderrs "errors"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/pkg/velez_api"
)

const configFetchingPostfix = "_config_scanning"

func (c *ContainerManager) FetchConfig(ctx context.Context, req *velez_api.FetchConfig_Request) error {
	_, err := dockerutils.PullImage(ctx, c.docker, req.ImageName, false)
	if err != nil {
		return errors.Wrap(err, "error pulling image")
	}

	createReq := &velez_api.CreateSmerd_Request{
		Name:      req.ServiceName + configFetchingPostfix,
		ImageName: req.ImageName,
		Settings:  &velez_api.Container_Settings{},
	}

	cont, err := c.containerLauncher.createSimple(ctx, createReq)
	if err != nil {
		return errors.Wrap(err, "error creating container")
	}
	defer func() {
		_, dropErr := c.DropSmerds(ctx, &velez_api.DropSmerd_Request{
			Uuids: []string{cont.ID},
		})
		if dropErr != nil {
			err = stderrs.Join(err, errors.Wrap(dropErr, "error dropping config scanning smerd"))
		}
	}()

	configFromContainer, err := c.configManager.GetFromContainer(ctx, cont.ID)
	if err != nil {
		return errors.Wrap(err, "error getting matreshka config from container")
	}

	configFromApi, err := c.configManager.GetFromApi(ctx, req.GetServiceName())
	if err != nil {
		return errors.Wrap(err, "error getting matreshka config from matreshka api")
	}

	matreshkaConfig := matreshka.MergeConfigs(configFromApi, configFromContainer)

	err = c.configManager.UpdateConfig(ctx, req.ServiceName, matreshkaConfig)
	if err != nil {
		return errors.Wrap(err, "error updating config")
	}

	return nil
}
