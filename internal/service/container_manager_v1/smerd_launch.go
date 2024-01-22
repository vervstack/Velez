package container_manager_v1

import (
	"context"
	"fmt"
	"strings"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils/parser"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/pkg/velez_api"
)

const vervName = "VERV_NAME"

func (c *containerManager) LaunchSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (*velez_api.Smerd, error) {
	image, err := c.getImage(ctx, req.ImageName)
	if err != nil {
		return nil, errors.Wrap(err, "error getting image")
	}

	if req.Name == "" {
		req.Name = strings.Split(image.Name, "/")[1]
	}

	cfg := &container.Config{
		Image:    image.Name,
		Hostname: req.Name,
		Cmd:      parser.FromCommand(req.Command),
	}

	err = c.fillPorts(req.Settings.Ports)
	if err != nil {
		return nil, errors.Wrap(err, "error getting ports on host side")
	}

	cfg.Env, err = c.getConfigEnvs(ctx, req.Name)
	if err != nil {
		return nil, errors.Wrap(err, "error getting envs from config")
	}

	cfg.Env = append(cfg.Env,
		fmt.Sprintf("%s=%s", vervName, req.Name),
	)

	serviceContainer, err := c.docker.ContainerCreate(ctx,
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
		return nil, errors.Wrap(err, "error creating container")
	}

	err = c.docker.ContainerStart(ctx, serviceContainer.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error starting container")
	}

	out := &velez_api.Smerd{
		Uuid:      serviceContainer.ID,
		ImageName: req.ImageName,
	}

	cl, err := c.docker.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.Arg("id", serviceContainer.ID)),
	})

	if err == nil && len(cl) == 1 && len(cl[0].Names) > 0 {
		out.Name = cl[0].Names[0][1:]
	}

	return out, nil
}

func (c *containerManager) getConfigEnvs(ctx context.Context, name string) ([]string, error) {
	confRaw, err := c.matreshkaClient.GetConfigRaw(ctx,
		&matreshka_api.GetConfigRaw_Request{
			ServiceName: name,
		},
	)
	if err != nil {
		logrus.Warnf("error getting config for service \"%s\". Error: %s", name, err)
		return nil, nil
	}

	if confRaw == nil || confRaw.Config == "" {
		logrus.Warnf("no config returned for service \"%s\"", name)
		return nil, nil
	}

	conf, err := matreshka.ParseConfig([]byte(confRaw.Config))
	if err != nil {
		return nil, errors.Wrap(err, "error parsing config")
	}

	envKeys, err := matreshka.GenerateKeys(conf)
	if err != nil {
		return nil, errors.Wrap(err, "error generating keys")
	}

	keys := make([]string, 0, len(envKeys))

	for _, item := range envKeys {
		keys = append(keys, item.Name+"="+fmt.Sprintf("%v", item.Value))
	}

	return keys, nil
}

func (c *containerManager) fillPorts(req []*velez_api.PortBindings) error {
	for i := range req {
		port := c.portManager.GetPort()
		if port == nil {
			return service.ErrNoPortsAvailable
		}

		req[i].Host = uint32(*port)
	}

	return nil
}
