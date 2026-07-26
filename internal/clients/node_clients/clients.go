package node_clients

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/local_state"
	"go.vervstack.ru/Velez/internal/clients/node_clients/ports"
	"go.vervstack.ru/Velez/internal/domain"
)

// PortManager is an alias for the interface defined in the ports sub-package.
type PortManager = ports.PortManager

type Docker interface {
	PullImage(ctx context.Context, imageName string) (image.InspectResponse, error)
	Remove(ctx context.Context, uuid string) error
	Stop(ctx context.Context, nameOrId string) error
	Restart(ctx context.Context, nameOrId string) error
	ListContainers(ctx context.Context, req *velez_api.ListSmerds_Request) ([]container.Summary, error)
	ListOccupiedPorts(ctx context.Context) ([]uint32, error)

	Exec(ctx context.Context, contId string, options container.ExecOptions) ([]byte, error)

	// IsContainerRunning returns (running=true, exists=true) if the container is up,
	// (running=false, exists=true) if it exists but has exited/paused/died,
	// and (false, false, nil) if no container with that name/id was found.
	IsContainerRunning(ctx context.Context, nameOrId string) (running bool, exists bool, err error)

	Client() client.APIClient

	ContainerCreate(
		ctx context.Context,
		config *container.Config,
		hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig,
		platform *ocispec.Platform,
		containerName string,
	) (container.CreateResponse, error)

	Stats(ctx context.Context, nameOrId string) (domain.ContainerStats, error)
}

type StateManager interface {
	Start() error
	Stop() error

	Set(st local_state.State)
	Get() local_state.State

	GetForUpdate() local_state.State
	SetAndRelease(state local_state.State)

	ValidateVelezPrivateKey(in string) bool
}

type HardwareManager interface {
	GetHardware() (*velez_api.GetHardware_Response, error)
}
