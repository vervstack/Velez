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

//nolint:interfacebloat
type Docker interface {
	PullImage(ctx context.Context, imageName string) (image.InspectResponse, error)
	Remove(ctx context.Context, uuid string) error
	Stop(ctx context.Context, nameOrId string) error
	Restart(ctx context.Context, nameOrId string) error
	// ListContainers lists containers scoped to the environment identified by
	// suffix (labels.SuffixLabel). An empty suffix lists across every
	// environment on the node.
	ListContainers(
		ctx context.Context, req *velez_api.ListSmerds_Request, suffix string,
	) ([]container.Summary, error)
	ListOccupiedPorts(ctx context.Context) ([]uint32, error)

	Exec(ctx context.Context, contId string, options container.ExecOptions) ([]byte, error)

	// IsContainerRunning returns (running=true, exists=true) if the container is up,
	// (running=false, exists=true) if it exists but has exited/paused/died,
	// and (false, false, nil) if no container with that name/id was found.
	IsContainerRunning(ctx context.Context, nameOrId string) (running bool, exists bool, err error)

	Client() client.APIClient

	// Host is the resolved address of the Docker daemon this connection talks
	// to (github.com/docker/docker/client.Client.DaemonHost - DOCKER_HOST when
	// set, else the SDK's platform default socket). It's the default
	// domain.Environment.DockerHost an environment gets when none is given
	// explicitly - see verv_services.CreateEnvironment.
	Host() string

	ContainerCreate(
		ctx context.Context,
		config *container.Config,
		hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig,
		platform *ocispec.Platform,
		containerName string,
		// suffix - resolved environment suffix, stamped onto the container as
		// labels.SuffixLabel.
		suffix string,
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
