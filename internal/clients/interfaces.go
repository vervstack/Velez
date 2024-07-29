package clients

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/godverv/makosh/pkg/makosh_be"
	"github.com/godverv/matreshka"

	"github.com/godverv/Velez/pkg/velez_api"
)

type InternalClients interface {
	Docker() Docker
	DeployManager() DeployManager
	PortManager() PortManager
	HardwareManager() HardwareManager

	SecurityManager() SecurityManager
}

type ExternalClients interface {
	//ServiceDiscovery() ServiceDiscovery
	Configurator() Configurator
}

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

type Configurator interface {
	GetFromContainer(ctx context.Context, contId string) (matreshka.AppConfig, error)
	GetFromApi(ctx context.Context, serviceName string) (matreshka.AppConfig, error)
	UpdateConfig(ctx context.Context, serviceName string, config matreshka.AppConfig) error
}

type PortManager interface {
	GetPort() (uint32, error)
	LockPort(ports ...uint32) error
	UnlockPorts(ports []uint32)
}

type HardwareManager interface {
	GetHardware() (*velez_api.GetHardware_Response, error)
}

type ServiceDiscovery interface {
	GetToken() string
	makosh_be.MakoshBeAPIClient
}

type SecurityManager interface {
	Start() error
	Stop() error

	ValidateKey(in string) bool
}
