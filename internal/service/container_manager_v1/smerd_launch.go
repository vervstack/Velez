package container_manager_v1

import (
	"context"
	"fmt"
	"strings"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/godverv/Velez/internal/backservice/configuration"
	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils/parser"
	"github.com/godverv/Velez/internal/utils/common"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	vervName  = "VERV_NAME"
	configURL = "CONFIG"
)

func (c *ContainerManager) LaunchSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (id string, err error) {
	image, err := c.getImage(ctx, req.ImageName)
	if err != nil {
		return "", errors.Wrap(err, "error getting image")
	}

	if req.Name == "" {
		req.Name = strings.Split(image.Name, "/")[1]
	}

	cfg, err := c.getCreateConfig(ctx, req)
	if err != nil {
		return "", errors.Wrap(err, "error getting create config")
	}

	if common.Contains(image.Tags, configURL) {
		cfg.Env = append(cfg.Env, configURL+"="+configuration.Name)
	}

	cont, err := c.docker.ContainerCreate(ctx,
		cfg,
		&container.HostConfig{
			PortBindings: parser.FromPorts(req.Settings),
			Mounts:       parser.FromBind(req.Settings),
		},
		&network.NetworkingConfig{},
		&v1.Platform{},
		req.Name,
	)
	if err != nil {
		return "", errors.Wrap(err, "error creating container")
	}

	err = c.docker.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", errors.Wrap(err, "error starting container")
	}

	err = c.docker.NetworkConnect(ctx,
		env.VervNetwork,
		cont.ID,
		&network.EndpointSettings{Aliases: []string{req.Name}},
	)
	if err != nil {
		return "", errors.Wrap(err, "error connecting container to verv network")
	}

	return cont.ID, nil
}

func (c *ContainerManager) getCreateConfig(ctx context.Context, req *velez_api.CreateSmerd_Request) (cfg *container.Config, err error) {
	cfg = &container.Config{
		Image:    req.ImageName,
		Hostname: req.Name,
		Cmd:      parser.FromCommand(req.Command),
	}

	if req.Settings != nil {
		err = c.portManager.GetPorts(req.Settings.Ports)
		if err != nil {
			return nil, errors.Wrap(err, "error getting ports on host side")
		}
	}

	cfg.Env, err = c.configManager.GetConfigEnvs(ctx, req.Name)
	if err != nil {
		return nil, errors.Wrap(err, "error getting envs from config")
	}

	cfg.Env = append(cfg.Env,
		fmt.Sprintf("%s=%s", vervName, req.Name),
	)

	return cfg, nil
}
