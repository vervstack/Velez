package container_manager_v1

import (
	"context"
	"fmt"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka/resources"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils/parser"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1/resource_manager"
	"github.com/godverv/Velez/pkg/velez_api"
)

// TODO VERV-43: use configurator from verv in node mode
// when resolver will be released - migrate over to resolver
const matreshkaUrl = "matreshka:80"

func (c *ContainerManager) createSimpleContainer(ctx context.Context, cfg *container.Config, req *velez_api.CreateSmerd_Request) (*types.Container, error) {
	var hostConfig container.HostConfig
	{
		hostConfig.PortBindings = parser.FromPorts(req.Settings)
		hostConfig.Mounts = parser.FromBind(req.Settings)
		if len(req.Settings.Volumes) != 0 {
			hostConfig.VolumeDriver = req.Settings.Volumes[0].Volume
		}
	}

	cont, err := c.docker.ContainerCreate(ctx,
		cfg,
		&hostConfig,
		&network.NetworkingConfig{},
		&v1.Platform{},
		req.GetName(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating container")
	}

	cl, err := dockerutils.ListContainers(ctx,
		c.docker,
		&velez_api.ListSmerds_Request{Id: &cont.ID})
	if err != nil {
		return nil, errors.Wrap(err, "error listing container by id")
	}

	for _, item := range cl {
		if item.ID == cont.ID {
			return &item, nil
		}
	}

	return nil, errors.Wrap(err, "container was created but never found")
}

func (c *ContainerManager) createVervContainer(ctx context.Context, cfg *container.Config, req *velez_api.CreateSmerd_Request) (*types.Container, error) {
	// Create network for smerd
	{
		err := dockerutils.CreateNetworkSoft(ctx, c.docker, req.GetName())
		if err != nil {
			return nil, errors.Wrap(err, "error creating network for service")
		}
		req.Settings.Networks = append(req.Settings.Networks,
			&velez_api.NetworkBind{
				NetworkName: req.GetName(),
				Aliases:     []string{req.GetName()},
			})
	}

	// Verv-Env variables
	{
		cfg.Env = append(cfg.Env,
			fmt.Sprintf("%s=%s", matreshka.VervName, req.GetName()),
			fmt.Sprintf("%s=%s", matreshka.ApiURL, matreshkaUrl))
	}

	cont, err := c.createSimpleContainer(ctx, cfg, req)
	if err != nil {
		return nil, errors.Wrap(err, "error creating pre container")
	}

	var matreshkaConfig matreshka.AppConfig
	{
		var configFromContainer matreshka.AppConfig
		configFromContainer, err = c.configManager.GetFromContainer(ctx, cont.ID)
		if err != nil {
			return nil, errors.Wrap(err, "error getting matreshka config from container")
		}
		var configFromApi matreshka.AppConfig
		configFromApi, err = c.configManager.GetFromApi(ctx, req.GetName())
		if err != nil {
			return nil, errors.Wrap(err, "error getting matreshka config from matreshka api")
		}

		matreshkaConfig = matreshka.MergeConfigs(configFromApi, configFromContainer)
	}

	if c.isNodeModeOn {
		err = c.setupDependencies(ctx, req, &matreshkaConfig)
		if err != nil {
			return nil, errors.Wrap(err, "error setting up dependencies")
		}
	}
	err = c.configManager.UpdateConfig(ctx, req.GetName(), matreshkaConfig)
	if err != nil {
		return nil, errors.Wrap(err, "error updating config")
	}

	return cont, nil
}

func (c *ContainerManager) setupDependencies(
	ctx context.Context,
	req *velez_api.CreateSmerd_Request,
	matreshkaCfg *matreshka.AppConfig,
) (err error) {
	images := make([]*velez_api.CreateSmerd_Request, 0, len(matreshkaCfg.Resources))
	for _, cfgResource := range matreshkaCfg.Resources {
		tp := cfgResource.GetType()
		switch tp {
		case resources.GrpcResourceName, resources.TelegramResourceName:
			continue
		}

		var constructor resource_manager.SmerdConstructor
		constructor, err = c.resourceManager.GetByName(tp)
		if err != nil {
			return errors.Wrap(err, "error getting name for resource")
		}

		var createSmerdRequest *velez_api.CreateSmerd_Request
		createSmerdRequest, err = constructor(matreshkaCfg.Resources, cfgResource.GetName())
		if err != nil {
			return errors.Wrap(err, "error getting resource-smerd config ")
		}

		name := req.GetName() + "_" + cfgResource.GetName()
		createSmerdRequest.Name = &name

		createSmerdRequest.Settings.Networks = append(createSmerdRequest.Settings.Networks,
			&velez_api.NetworkBind{
				NetworkName: req.GetName(),
				Aliases:     []string{cfgResource.GetName()},
			},
		)

		images = append(images, createSmerdRequest)
	}

	for _, resourceSmerd := range images {
		_, err = c.LaunchSmerd(ctx, resourceSmerd)
		if err != nil {
			return errors.Wrap(err, "error launching resource smerd")
		}
	}

	return nil
}
