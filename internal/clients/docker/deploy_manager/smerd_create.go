package deploy_manager

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	errors "go.redsock.ru/rerrors"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (d *DeployManager) Create(ctx context.Context, req *velez_api.CreateSmerd_Request) (*types.ContainerJSON, error) {
	cfg := getLaunchConfig(req)
	hCfg := getHostConfig(req)
	nCfg := getNetworkConfig(req)
	pCfg := &v1.Platform{}

	cont, err := d.docker.ContainerCreate(ctx, cfg, hCfg, nCfg, pCfg, req.GetName())
	if err != nil {
		return nil, errors.Wrap(err, "error creating container")
	}

	cl, err := d.docker.ContainerInspect(ctx, cont.ID)
	if err != nil {
		return nil, errors.Wrap(err, "error inspecting container by id")
	}

	err = d.docker.ContainerStart(ctx, cont.ID, container.StartOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error starting container")
	}

	return &cl, nil
}
