package container_manager_v1

import (
	"context"
	"fmt"
	"strings"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"

	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils/parser"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	vervName                         = "VERV_NAME"
	defaultContainerConfigFolderPath = "/app/config"
	matreshkaConfigLabel             = "MATRESHKA_CONFIG_ENABLED"
)

func (c *ContainerManager) LaunchSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (id string, err error) {
	c.normalizeCreateRequest(req)

	image, err := dockerutils.PullImage(ctx, c.docker, req.ImageName, false)
	if err != nil {
		return "", errors.Wrap(err, "error pulling image")
	}

	cfg, err := c.getLaunchConfig(req, image)
	if err != nil {
		return "", errors.Wrap(err, "error during launching config creation")
	}

	if req.Settings != nil {
		err = c.portManager.FillPorts(req.Settings.Ports)
		if err != nil {
			return "", errors.Wrap(err, "error getting ports on host side")
		}
	}

	var cont *types.Container

	if image.Labels[matreshkaConfigLabel] == "true" {
		cont, err = c.createVervContainer(ctx, cfg, req)
	} else {
		var envVars []string
		envVars, err = c.configManager.GetEnv(ctx, req.GetName())
		if err != nil {
			return "", errors.Wrap(err, "error obtaining config for container environment")
		}

		cfg.Env = append(cfg.Env, envVars...)

		cont, err = c.createSimpleContainer(ctx, cfg, req)
	}
	if err != nil {
		return "", errors.Wrap(err, "error creating container")
	}

	req.Settings.Networks = append(req.Settings.Networks,
		&velez_api.NetworkBind{
			NetworkName: env.VervNetwork,
			Aliases:     []string{req.GetName()},
		})

	for networkName, networkSettings := range parser.FromNetwork(req.Settings) {
		err = c.docker.NetworkConnect(ctx,
			networkName,
			cont.ID,
			networkSettings,
		)
		if err != nil {
			return "", errors.Wrap(err, "error connecting container to verv network")
		}
	}

	err = c.docker.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", errors.Wrap(err, "error starting container")
	}

	return cont.ID, nil
}

func (c *ContainerManager) getLaunchConfig(req *velez_api.CreateSmerd_Request, image *velez_api.Image) (cfg *container.Config, err error) {
	if req.Name == nil {
		req.Name = &strings.Split(image.Name, "/")[1]
	}

	cfg = &container.Config{
		Image:    req.ImageName,
		Hostname: req.GetName(),
		Cmd:      parser.FromCommand(req.Command),
		Env: []string{
			fmt.Sprintf("%s=%s", vervName, req.GetName()),
		},
	}

	for k, v := range req.Env {
		cfg.Env = append(cfg.Env, k+"="+v)
	}

	return cfg, nil
}

func (c *ContainerManager) normalizeCreateRequest(req *velez_api.CreateSmerd_Request) {
	if req.Settings == nil {
		req.Settings = &velez_api.Container_Settings{}
	}

	if req.Hardware == nil {
		req.Hardware = &velez_api.Container_Hardware{}
	}
}
