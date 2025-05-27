package clients

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"go.vervstack.ru/makosh/pkg/makosh_be"
	"go.vervstack.ru/matreshka/pkg/matreshka_be_api"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

type Docker interface {
	PullImage(ctx context.Context, imageName string) (image.InspectResponse, error)
	Remove(ctx context.Context, uuid string) error
	ListContainers(ctx context.Context, req *velez_api.ListSmerds_Request) ([]container.Summary, error)
	InspectContainer(ctx context.Context, containerID string) (container.InspectResponse, error)
	InspectImage(ctx context.Context, image string) (image.InspectResponse, error)

	client.APIClient
}

type PortManager interface {
	GetPort() (uint32, error)
	LockPort(ports ...uint32) error
	UnlockPorts(ports []uint32)

	// HoldPort returns true if port was not on hold (you set it on hold)
	// returns false if port is already on hold
	HoldPort(port uint32) bool
	// UnHoldPort return true if port was on hold (you take it)
	// returns false if port is not on hold
	UnHoldPort(port uint32) bool
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
