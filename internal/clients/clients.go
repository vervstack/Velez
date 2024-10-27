package clients

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/godverv/makosh/pkg/makosh_be"
	"github.com/godverv/matreshka-be/pkg/matreshka_be_api"

	"github.com/godverv/Velez/pkg/velez_api"
)

type Docker interface {
	PullImage(ctx context.Context, imageName string) (types.ImageInspect, error)
	Remove(ctx context.Context, uuid string) error
	ListContainers(ctx context.Context, req *velez_api.ListSmerds_Request) ([]types.Container, error)
	InspectContainer(ctx context.Context, containerID string) (types.ContainerJSON, error)
	InspectImage(ctx context.Context, image string) (types.ImageInspect, error)

	client.CommonAPIClient
}

type DeployManager interface {
	Create(ctx context.Context, req *velez_api.CreateSmerd_Request) (*types.ContainerJSON, error)
	Healthcheck(ctx context.Context, contId string, healthcheck *velez_api.Container_Healthcheck) error
}

type PortManager interface {
	GetPort() (uint32, error)
	LockPort(ports ...uint32) error
	UnlockPorts(ports []uint32)
}

type SecurityManager interface {
	Start() error
	Stop() error

	ValidateKey(in string) bool
}

type HardwareManager interface {
	GetHardware() (*velez_api.GetHardware_Response, error)
}

type Configurator interface {
	matreshka_be_api.MatreshkaBeAPIClient
}

type ServiceDiscovery interface {
	makosh_be.MakoshBeAPIClient
}
