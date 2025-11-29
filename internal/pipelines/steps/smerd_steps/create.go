package smerd_steps

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/backservice/env"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/domain"
)

type createSmerdStep struct {
	docker    node_clients.Docker
	dockerAPI client.APIClient

	req         *domain.LaunchSmerd
	containerId *string
}

func Create(nodeClients node_clients.NodeClients,
	req *domain.LaunchSmerd,
	containerId *string,
) *createSmerdStep {
	return &createSmerdStep{
		docker:      nodeClients.Docker(),
		dockerAPI:   nodeClients.Docker().Client(),
		req:         req,
		containerId: containerId,
	}
}

func (s *createSmerdStep) Do(ctx context.Context) error {
	cfg := s.getLaunchConfig()
	hCfg := s.getHostConfig()
	nCfg := s.getNetworkConfig()
	pCfg := &v1.Platform{}

	contName := s.req.GetName()

	createdContainer, err := s.dockerAPI.ContainerCreate(ctx, cfg, hCfg, nCfg, pCfg, contName)
	if err != nil {
		return rerrors.Wrap(err, "error creating container")
	}

	containerInfo, err := s.dockerAPI.ContainerInspect(ctx, createdContainer.ID)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container by id")
	}

	*s.containerId = containerInfo.ID

	for _, n := range s.req.Settings.Network {
		connectReq := dockerutils.ConnectToNetworkRequest{
			NetworkName: n.NetworkName,
			ContId:      createdContainer.ID,
			Aliases:     n.Aliases,
		}
		err = dockerutils.ConnectToNetwork(ctx, s.dockerAPI, connectReq)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	return nil
}

func (s *createSmerdStep) Rollback(ctx context.Context) error {
	if s.containerId == nil {
		return nil
	}

	err := s.docker.Remove(ctx, *s.containerId)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return rerrors.Wrapf(err, "error removing container '%s'", *s.containerId)
		}
	}

	return nil
}

func (s *createSmerdStep) getLaunchConfig() (cfg *container.Config) {
	cfg = &container.Config{
		Image:       s.req.ImageName,
		Hostname:    s.req.GetName(),
		Cmd:         parser.FromCommand(s.req.Command),
		Healthcheck: parser.FromHealthcheck(s.req.Healthcheck),
		Env:         parser.FromDockerEnv(s.req.Env),
		Labels:      s.req.Labels,
	}

	return cfg
}

func (s *createSmerdStep) getHostConfig() (hostConfig *container.HostConfig) {
	hostConfig = &container.HostConfig{
		PortBindings:  parser.FromPorts(s.req.Settings),
		Mounts:        parser.FromVolume(s.req.Settings),
		RestartPolicy: parser.FromRestart(s.req.Restart),
	}

	if s.req.Settings != nil && len(s.req.Settings.Volumes) != 0 {
		hostConfig.VolumeDriver = s.req.Settings.Volumes[0].VolumeName
	}

	return hostConfig
}

func (s *createSmerdStep) getNetworkConfig() (networkConfig *network.NetworkingConfig) {
	networkConfig = &network.NetworkingConfig{}

	if len(s.req.Settings.Ports) == 0 {
		return networkConfig
	}

	networkConfig.EndpointsConfig = make(map[string]*network.EndpointSettings)

	vervNetwork := &network.EndpointSettings{
		Aliases: []string{s.req.GetName()},
	}
	networkConfig.EndpointsConfig[env.VervNetwork] = vervNetwork

	// required in order to expose ports on some platforms (e.g. orbs)
	networkConfig.EndpointsConfig["bridge"] = &network.EndpointSettings{}

	for _, v := range s.req.Settings.Network {
		networkConfig.EndpointsConfig[v.NetworkName] = &network.EndpointSettings{
			Aliases: v.Aliases,
		}
	}

	return networkConfig
}
