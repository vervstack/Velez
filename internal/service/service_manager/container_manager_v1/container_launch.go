package container_manager_v1

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka/resources"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils/parser"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1/config_manager"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1/resource_manager"
	"github.com/godverv/Velez/pkg/velez_api"
)

// TODO VERV-43: use configurator from verv in node mode
// when resolver will be released - migrate over to resolver
const matreshkaUrl = "matreshka:80"

type ContainerLauncher struct {
	docker client.CommonAPIClient

	configManager   *config_manager.Configurator
	resourceManager *resource_manager.ResourceManager

	isNodeModeOn bool
}

func (c *ContainerLauncher) createSimple(ctx context.Context, req *velez_api.CreateSmerd_Request) (*types.Container, error) {
	cfg := getLaunchConfig(req)
	hCfg := getHostConfig(req)
	nCfg := getNetworkConfig(req)
	pCfg := &v1.Platform{}

	cont, err := c.docker.ContainerCreate(ctx, cfg, hCfg, nCfg, pCfg, req.GetName())
	if err != nil {
		return nil, errors.Wrap(err, "error creating container")
	}

	listReq := &velez_api.ListSmerds_Request{Id: &cont.ID}
	cl, err := dockerutils.ListContainers(ctx, c.docker, listReq)
	if err != nil {
		return nil, errors.Wrap(err, "error listing container by id")
	}

	var createdContainer *types.Container
	for _, item := range cl {
		if item.ID == cont.ID {
			createdContainer = &item
			break
		}
	}

	if createdContainer == nil {
		return nil, errors.Wrap(err, "container was created but could not be found")
	}

	additionalNetworks := parser.FromNetworks(req.Settings)
	err = c.linkToNetworks(ctx, cont.ID, additionalNetworks)
	if err != nil {
		return nil, errors.Wrap(err, "error linking to networks")
	}

	return createdContainer, nil
}

func (c *ContainerLauncher) createVerv(ctx context.Context, req *velez_api.CreateSmerd_Request) (*types.Container, error) {
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
			},
		)
	}

	// Verv-Env variables
	{
		req.Env[matreshka.VervName] = req.GetName()
		req.Env[matreshka.ApiURL] = matreshkaUrl
	}

	cont, err := c.createSimple(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "error creating pre container")
	}

	// TODO VERV-52
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

func (c *ContainerLauncher) setupDependencies(
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

		createSmerdRequest.Name = req.GetName() + "_" + cfgResource.GetName()

		createSmerdRequest.Settings.Networks = append(createSmerdRequest.Settings.Networks,
			&velez_api.NetworkBind{
				NetworkName: req.GetName(),
				Aliases:     []string{cfgResource.GetName()},
			},
		)

		images = append(images, createSmerdRequest)
	}

	//for _, resourceSmerd := range images {
	//_, err = c.LaunchSmerd(ctx, resourceSmerd)
	//if err != nil {
	//	return errors.Wrap(err, "error launching resource smerd")
	//}
	//}

	return nil
}

func (c *ContainerLauncher) linkToNetworks(
	ctx context.Context, contId string, networks map[string]*network.EndpointSettings) error {
	for networkName, networkSettings := range networks {
		err := c.docker.NetworkConnect(ctx, networkName, contId, networkSettings)
		if err != nil {
			return errors.Wrap(err, "error connecting container to network")
		}
	}
	return nil
}

func getLaunchConfig(req *velez_api.CreateSmerd_Request) (cfg *container.Config) {
	cfg = &container.Config{
		Image:       req.ImageName,
		Hostname:    req.GetName(),
		Cmd:         parser.FromCommand(req.Command),
		Healthcheck: parser.FromHealthcheck(req.Healthcheck),
		Env:         make([]string, 0, len(req.Env)),
	}

	for k, v := range req.Env {
		cfg.Env = append(cfg.Env, k+"="+v)
	}

	return cfg
}

func getHostConfig(req *velez_api.CreateSmerd_Request) (hostConfig *container.HostConfig) {
	// TODO https://redsock.youtrack.cloud/issue/VERV-56
	hostConfig = &container.HostConfig{
		PortBindings: parser.FromPorts(req.Settings),
		Mounts:       parser.FromBind(req.Settings),
	}

	if req.Settings != nil && len(req.Settings.Volumes) != 0 {
		hostConfig.VolumeDriver = req.Settings.Volumes[0].Volume
	}

	return hostConfig
}

func getNetworkConfig(req *velez_api.CreateSmerd_Request) (networkConfig *network.NetworkingConfig) {
	networkConfig = &network.NetworkingConfig{}

	networkConfig.EndpointsConfig = make(map[string]*network.EndpointSettings)

	vervNetwork := &network.EndpointSettings{
		Aliases: []string{req.GetName()},
	}
	networkConfig.EndpointsConfig[env.VervNetwork] = vervNetwork

	return networkConfig
}
