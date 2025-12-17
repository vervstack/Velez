package node_clients

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"go.vervstack.ru/Velez/internal/clients/node_clients/state"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type Docker interface {
	PullImage(ctx context.Context, imageName string) (image.InspectResponse, error)
	Remove(ctx context.Context, uuid string) error
	ListContainers(ctx context.Context, req *velez_api.ListSmerds_Request) ([]container.Summary, error)

	Exec(ctx context.Context, contId string, options container.ExecOptions) ([]byte, error)

	Client() client.APIClient

	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
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

type StateManager interface {
	Start() error
	Stop() error

	Set(st state.State)
	Get() state.State

	ValidateVelezPrivateKey(in string) bool
}

type HardwareManager interface {
	GetHardware() (*velez_api.GetHardware_Response, error)
}
