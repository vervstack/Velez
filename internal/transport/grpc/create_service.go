package grpc

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Api) CreateSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (*velez_api.CreateSmerd_Response, error) {
	err := req.ValidateAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "invalid request. Must match \"lowercase/lowercase:v0.0.1\"").Error())
	}

	smerd, err := a.containerManager.LaunchSmerd(ctx, domain.ContainerCreate{
		Name:            req.Name,
		ImageName:       req.ImageName,
		AllowDuplicates: req.AllowDuplicates,
		Ports:           getPorts(req.Settings),
		Volumes:         getVolumes(req.Settings),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error searching image")
	}
	_ = smerd

	return &velez_api.CreateSmerd_Response{}, nil
}

func (a *Api) getContainerConfig(name string) *container.Config {
	return &container.Config{
		Image: name,
	}
}

func (a *Api) getContainerHostConfig() *container.HostConfig {
	return nil
}

func (a *Api) getNetworkConfig() *network.NetworkingConfig {
	return nil
}

func (a *Api) getPlatform() *v1.Platform {
	return nil
}
