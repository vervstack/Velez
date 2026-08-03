package container_runtime

import (
	"github.com/docker/docker/client"
)

// commonRuntime backs operations that don't differ by backend. It only needs a
// raw client.APIClient and doesn't know - or care - whether that connection
// points at the node's shared daemon or at one dedicated to a single
// environment.
//
// It carries no methods yet: Phase 1 only implements ContainerCreate, which is
// backend-specific and therefore lives on the concrete runtimes. The
// backend-agnostic operations (PullImage, CreateNetwork, ConnectToNetwork,
// DisconnectFromNetworks - see docs/container_runtimes/interface_design.md)
// land here, implemented once and embedded by every runtime, when their phase
// comes.
type commonRuntime struct {
	cli client.APIClient
}
