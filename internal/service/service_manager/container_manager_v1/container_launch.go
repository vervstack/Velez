package container_manager_v1

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/godverv/matreshka"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils/parser"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1/config_manager"
	"github.com/godverv/Velez/pkg/velez_api"
)

// TODO VERV-43: use configurator from verv in node mode
// when resolver will be released - migrate over to resolver
const (
	matreshkaUrl          = "matreshka:50050"
	CreatedWithVelezLabel = "CREATED_WITH_VELEZ"
)

type ContainerStarter struct {
	docker client.CommonAPIClient

	configManager *config_manager.Configurator
	isNodeModeOn  bool
}

func (c *ContainerStarter) createSimple(ctx context.Context, req *velez_api.CreateSmerd_Request) (*types.ContainerJSON, error) {
	cfg := getLaunchConfig(req)
	hCfg := getHostConfig(req)
	nCfg := getNetworkConfig(req)
	pCfg := &v1.Platform{}

	cont, err := c.docker.ContainerCreate(ctx, cfg, hCfg, nCfg, pCfg, req.GetName())
	if err != nil {
		return nil, errors.Wrap(err, "error creating container")
	}

	cl, err := c.docker.ContainerInspect(ctx, cont.ID)
	if err != nil {
		return nil, errors.Wrap(err, "error listing container by id")
	}

	return &cl, nil
}

func (c *ContainerStarter) createVerv(ctx context.Context, req *velez_api.CreateSmerd_Request) (*types.ContainerJSON, error) {
	matreshkaConfig, err := c.configManager.GetFromApi(ctx, req.GetName())
	if err != nil {
		return nil, errors.Wrap(err, "error getting matreshka config from matreshka api")
	}

	for _, srv := range matreshkaConfig.Servers {
		req.Settings.Ports = append(req.Settings.Ports,
			&velez_api.PortBindings{
				Container: uint32(srv.GetPort()),
				Protoc:    velez_api.PortBindings_tcp,
			})
	}

	// Create network for smerd
	{
		err = dockerutils.CreateNetworkSoft(ctx, c.docker, req.GetName())
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
		//req.Env[matreshka.ApiURL] = matreshkaUrl
	}

	cont, err := c.createSimple(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "error creating pre container")
	}

	return cont, nil
}

func getLaunchConfig(req *velez_api.CreateSmerd_Request) (cfg *container.Config) {
	req.Labels[CreatedWithVelezLabel] = "true"

	cfg = &container.Config{
		Image:       req.ImageName,
		Hostname:    req.GetName(),
		Cmd:         parser.FromCommand(req.Command),
		Healthcheck: parser.FromHealthcheck(req.Healthcheck),
		Env:         make([]string, 0, len(req.Env)),
		Labels:      req.Labels,
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
