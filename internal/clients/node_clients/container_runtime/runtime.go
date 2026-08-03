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
// Phase 1 (done) implements ContainerCreate, ListContainers, Remove, Rename,
// IsContainerRunning, Inspect, Stop, Restart, Stats and Exec, and
// create_smerd/drop_smerd/upgrade_smerd's rename/rollback jobs (plus
// container_manager.InspectSmerd and copy_to_volume.go's copyFileJob) all
// resolve through RuntimeResolver instead of calling node_clients.Docker
// directly. ListOccupiedPorts, PullImage and network ops still go through
// node_clients.Docker directly - see roadmap.md's "Explicitly deferred".
package container_runtime

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
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

// ConnectToNetworkRequest bundles what ConnectToNetwork needs. ContainerID is
// the LOGICAL identifier (uuid or logical/Docker name) - resolved via the
// runtime's ownership check, same as every other identifier-taking method on
// this interface (Stop/Restart/Remove/Rename/...). NetworkName is the LOGICAL
// network name - suffixed by the runtime the same way CreateNetwork suffixes
// one, so callers only ever deal in logical names, never the real/suffixed
// Docker network name.
type ConnectToNetworkRequest struct {
	ContainerID string
	NetworkName string
	Aliases     []string
}

// ContainerRuntime is what the service/jobs layers depend on instead of a
// concrete Docker client.
//
// Phase 1 exposes ContainerCreate, ListContainers, Remove, Rename,
// IsContainerRunning, Inspect, Stop, Restart, Stats and Exec. The rest of the
// target shape (ListOccupiedPorts, plus the backend-agnostic
// PullImage/network operations) is in
// docs/container_runtimes/interface_design.md; those methods stay on
// node_clients.Docker until their own phase.
//
//nolint:interfacebloat
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

	// Rename renames a container identified by uuid or logical/Docker name to
	// newName, strictly scoped to the environment this runtime instance was
	// resolved for - same ownership/resolution semantics as Remove (a
	// container belonging to a different environment, or no container at all
	// under either identifier form, is treated as "nothing to rename" and
	// reported as success). newName is itself run through the environment's
	// suffixing before being handed to the backend, so callers only ever deal
	// in logical/virtual names.
	Rename(ctx context.Context, identifier, newName string) error

	// IsContainerRunning reports whether the container identified by uuid or
	// logical/Docker name is currently running, scoped to the environment
	// this runtime instance was resolved for. exists is false - with running
	// also false and err nil - both when no container exists under either
	// identifier form and when a container exists but belongs to a different
	// environment (same cross-environment protection as Remove/Rename).
	IsContainerRunning(ctx context.Context, nameOrID string) (running, exists bool, err error)

	// Inspect returns the container identified by uuid or logical/Docker name,
	// scoped to the environment this runtime instance was resolved for - same
	// resolution/ownership semantics as Remove/Rename/IsContainerRunning: a
	// container belonging to a different environment, or not found under
	// either identifier form, is reported as found=false (with no error), not
	// as an error. The returned InspectResponse's Name is rewritten to the
	// virtual/logical name, never the suffixed Docker name - see
	// labelBasedRuntime.Inspect.
	Inspect(ctx context.Context, identifier string) (container.InspectResponse, bool, error)

	// Stop stops the container identified by uuid or logical/Docker name,
	// scoped to the environment this runtime instance was resolved for - same
	// resolution/ownership semantics as Remove/Rename: a container belonging
	// to a different environment, or not found under either identifier form,
	// is treated as "nothing to stop" and reported as success, not an error.
	Stop(ctx context.Context, nameOrID string) error

	// Restart restarts the container identified by uuid or logical/Docker
	// name, scoped to the environment this runtime instance was resolved for
	// - same resolution/ownership semantics as Stop/Remove/Rename: a
	// container belonging to a different environment, or not found under
	// either identifier form, is treated as "nothing to restart" and reported
	// as success, not an error.
	Restart(ctx context.Context, nameOrID string) error

	// Stats returns the container identified by uuid or logical/Docker name's
	// current resource usage, scoped to the environment this runtime instance
	// was resolved for. Unlike Stop/Restart/Remove, there is no sensible
	// zero-value "success" for a container that doesn't resolve under either
	// identifier form or belongs to a different environment - that case
	// returns a non-nil error instead of a zero domain.ContainerStats.
	Stats(ctx context.Context, nameOrID string) (domain.ContainerStats, error)

	// Exec runs cfg inside the container identified by uuid or logical/Docker
	// name, scoped to the environment this runtime instance was resolved for
	// - same resolution/ownership semantics as Stats: a container that
	// doesn't resolve under either identifier form, or belongs to a
	// different environment, is a real error here, not the idempotent
	// success Stop/Restart/Remove/Rename report - there is no sensible
	// zero-value "success" for exec output against a container that isn't
	// there.
	Exec(ctx context.Context, containerID string, cfg container.ExecOptions) ([]byte, error)

	// CreateNetwork ensures a Docker bridge network exists for the LOGICAL
	// name given, suffixed to the environment this runtime instance was
	// resolved for - see labelBasedRuntime's networkName - so each
	// environment gets its own dedicated network instead of every
	// environment sharing one. Idempotent: does nothing if the (suffixed)
	// network already exists.
	CreateNetwork(ctx context.Context, name string) error

	// ConnectToNetwork connects a container to a network, both scoped to the
	// environment this runtime instance was resolved for: the container
	// identifier is resolved via the same ownership check as
	// Stop/Restart/Remove/Rename/Exec, and req.NetworkName is suffixed the
	// same way CreateNetwork suffixes a network name. Unlike
	// Stop/Restart/Remove/Rename - which treat a missing/foreign container as
	// idempotent success - and like Exec/Stats, a container not found under
	// either identifier form, or belonging to a different environment, is a
	// real error here: there's no sensible "connected" outcome for a
	// container that isn't there.
	ConnectToNetwork(ctx context.Context, req ConnectToNetworkRequest) error

	// DisconnectFromNetworks disconnects a container from each named network,
	// both scoped to the environment this runtime instance was resolved for -
	// same identifier resolution/ownership semantics as ConnectToNetwork, and
	// the same networkName suffixing applied to every entry in networks
	// before it's handed to the backend.
	DisconnectFromNetworks(ctx context.Context, containerID string, networks []string) error
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
