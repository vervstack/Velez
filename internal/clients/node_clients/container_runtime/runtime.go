// Package container_runtime is the pluggable "how do we actually talk to
// containers" layer between Velez's service/jobs layers and a concrete
// container backend.
//
// Which concrete implementation serves a given environment - the shared Docker
// daemon scoped by suffix labels/names (tier 1, the default), a Docker daemon
// dedicated to that environment (tier 2), or another backend entirely
// (podman/containerd/Kata) - is resolved per call by a RuntimeResolver and
// invisible to callers.
//
// Design record: docs/container_runtimes/{README,interface_design,roadmap}.md.
// Phase 1 (current) implements ContainerCreate, ListContainers and Remove;
// every other container operation still goes through node_clients.Docker
// directly.
package container_runtime

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// ContainerConfig, HostConfig and NetworkingConfig are thin pass-throughs over
// the Docker SDK's own types.
//
// Keeping bare SDK types in the interface would block any future non-Docker
// backend from satisfying it; translating to fully independent domain types
// now would be a speculative refactor for a capability that doesn't exist yet.
// Wrapping is the middle ground: callers already code against the Velez-owned
// name, so when a future backend needs a genuinely different shape, these
// wrappers absorb that change without touching every call site.
type ContainerConfig struct {
	*container.Config
}

type HostConfig struct {
	*container.HostConfig
}

type NetworkingConfig struct {
	*network.NetworkingConfig
}

// ContainerCreateRequest bundles what today's Docker.ContainerCreate takes as
// positional parameters. Note what's gone: no suffix parameter - the resolved
// environment is baked into the runtime instance the resolver hands out, not
// re-supplied by every caller.
type ContainerCreateRequest struct {
	Config           *ContainerConfig
	HostConfig       *HostConfig
	NetworkingConfig *NetworkingConfig
	Platform         *v1.Platform
	// ContainerName is the LOGICAL name (the smerd name). Turning it into the
	// actual Docker container name - including any environment suffixing - is
	// the runtime implementation's job.
	ContainerName string
}

// ContainerRuntime is what the service/jobs layers depend on instead of a
// concrete Docker client.
//
// Phase 1 exposes ContainerCreate and ListContainers. The target shape
// (Remove, Stop, Restart, IsContainerRunning, Exec, Stats, ListOccupiedPorts,
// plus the backend-agnostic PullImage/network operations) is in
// docs/container_runtimes/interface_design.md; those methods stay on
// node_clients.Docker until their own phase.
type ContainerRuntime interface {
	ContainerCreate(ctx context.Context, req ContainerCreateRequest) (container.CreateResponse, error)

	// ListContainers lists the containers belonging to the environment this
	// runtime instance was resolved for. Unlike node_clients.Docker's
	// ListContainers, there is no suffix parameter - the resolved
	// environment's suffix is baked into the runtime instance the resolver
	// handed out, not re-supplied per call. Every returned container.Summary's
	// Names are virtual/logical names, never the suffixed Docker name - see
	// labelBasedRuntime.ListContainers.
	ListContainers(ctx context.Context, req *velez_api.ListSmerds_Request) ([]container.Summary, error)

	// Remove deletes a single container identified by uuid or logical/Docker
	// name, strictly scoped to the environment this runtime instance was
	// resolved for: a container belonging to a different environment (or no
	// container at all, under either identifier form) is treated as "nothing
	// to remove" and reported as success, not an error - see
	// labelBasedRuntime.Remove's doc comment for the resolution/ownership
	// check this relies on.
	Remove(ctx context.Context, identifier string) error
}

// RuntimeResolver hands out the ContainerRuntime serving a given environment.
//
// Implementations re-read the environment's row on every call, so flipping an
// environment between tiers takes effect on the next call - no explicit
// atomic-swap container needed, unlike the state axis's static->pg swap (which
// has to move a process-wide singleton rather than a per-environment lookup).
type RuntimeResolver interface {
	Runtime(ctx context.Context, environment string) (ContainerRuntime, error)
}
