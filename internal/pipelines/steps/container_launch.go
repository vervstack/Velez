package steps

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils/parser"
	"github.com/godverv/Velez/pkg/velez_api"
)

type createContainerStep struct {
	docker clients.Docker

	req         *velez_api.CreateSmerd_Request
	containerId *string
}

func LaunchContainer(docker clients.Docker,
	req *velez_api.CreateSmerd_Request,
	containerId *string,
) *createContainerStep {
	return &createContainerStep{
		docker:      docker,
		req:         req,
		containerId: containerId,
	}
}

func (s *createContainerStep) Do(ctx context.Context) error {
	cfg := s.getLaunchConfig()
	hCfg := s.getHostConfig()
	nCfg := s.getNetworkConfig()
	pCfg := &v1.Platform{}

	createdContainer, err := s.docker.ContainerCreate(ctx, cfg, hCfg, nCfg, pCfg, s.req.GetName())
	if err != nil {
		return rerrors.Wrap(err, "error creating container")
	}

	containerInfo, err := s.docker.ContainerInspect(ctx, createdContainer.ID)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container by id")
	}

	*s.containerId = containerInfo.ID

	err = s.docker.ContainerStart(ctx, createdContainer.ID, container.StartOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error starting container")
	}

	return nil
}

func (s *createContainerStep) Rollback(ctx context.Context) error {
	if s.containerId == nil {
		return nil
	}

	err := s.docker.Remove(ctx, *s.containerId)
	if err != nil {
		return rerrors.Wrapf(err, "error removing container '%s'", *s.containerId)
	}

	return nil
}

func (s *createContainerStep) getLaunchConfig() (cfg *container.Config) {
	cfg = &container.Config{
		Image:       s.req.ImageName,
		Hostname:    s.req.GetName(),
		Cmd:         parser.FromCommand(s.req.Command),
		Healthcheck: parser.FromHealthcheck(s.req.Healthcheck),
		Env:         make([]string, 0, len(s.req.Env)),
		Labels:      s.req.Labels,
	}

	for k, v := range s.req.Env {
		cfg.Env = append(cfg.Env, k+"="+v)
	}

	return cfg
}

func (s *createContainerStep) getHostConfig() (hostConfig *container.HostConfig) {
	hostConfig = &container.HostConfig{
		PortBindings: parser.FromPorts(s.req.Settings),
		Mounts:       parser.FromBind(s.req.Settings),
		RestartPolicy: container.RestartPolicy{
			Name:              container.RestartPolicyOnFailure,
			MaximumRetryCount: 3,
		},
	}

	if s.req.Settings != nil && len(s.req.Settings.Volumes) != 0 {
		hostConfig.VolumeDriver = s.req.Settings.Volumes[0].VolumeName
	}

	return hostConfig
}

func (s *createContainerStep) getNetworkConfig() (networkConfig *network.NetworkingConfig) {
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

	return networkConfig
}
