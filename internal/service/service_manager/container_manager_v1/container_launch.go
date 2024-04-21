package container_manager_v1

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/godverv/matreshka/resources"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils/parser"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1/resource_manager"
	"github.com/godverv/Velez/pkg/velez_api"
)

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
	// mount config
	{
		mountPoint, err := c.configManager.GetMountPoint(req.GetName())
		if err != nil {
			return nil, errors.Wrap(err, "error getting mount point for container")
		}

		v, err := c.docker.VolumeCreate(ctx, volume.CreateOptions{
			Name: req.GetName(),
			DriverOpts: map[string]string{
				"type":   "none",
				"device": mountPoint,
				"o":      "bind",
			},
		})
		if err != nil {
			return nil, errors.Wrap(err, "error creating volume")
		}

		req.Settings.Volumes = append(req.Settings.Volumes, &velez_api.VolumeBindings{
			Volume:        v.Name,
			ContainerPath: defaultContainerConfigFolderPath,
		})
	}

	// mount network
	{
		err := dockerutils.CreateNetworkSoft(ctx, c.docker, req.GetName())
		if err != nil {
			return nil, errors.Wrap(err, "error creating network for service")
		}
		req.Settings.Networks = append(req.Settings.Networks, &velez_api.NetworkBind{
			NetworkName: req.GetName(),
			Aliases:     []string{req.GetName()},
		})
	}

	cont, err := c.createSimpleContainer(ctx, cfg, req)
	if err != nil {
		return nil, errors.Wrap(err, "error creating pre container")
	}

	matreshkaCfg, err := c.configManager.Mount(ctx, req.GetName())
	if err != nil {
		return nil, errors.Wrap(err, "error mouting matreshka config")
	}

	if c.isNodeModeOn {
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
				return nil, errors.Wrap(err, "error getting name for resource")
			}

			var createSmerdRequest *velez_api.CreateSmerd_Request
			createSmerdRequest, err = constructor(matreshkaCfg.Resources, cfgResource.GetName())
			if err != nil {
				return nil, errors.Wrap(err, "error getting resource-smerd config ")
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

		err = c.configManager.UpdateConfig(ctx, req.GetName(), matreshkaCfg)
		if err != nil {
			return nil, errors.Wrap(err, "error updating config")
		}

		for _, resourceSmerd := range images {
			_, err = c.LaunchSmerd(ctx, resourceSmerd)
			if err != nil {
				return nil, errors.Wrap(err, "error launching resource smerd")
			}
		}
	}

	return cont, nil
}
