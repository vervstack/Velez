package container_runtime

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	labelValueTrue = "true"

	// nameSuffixSeparator joins a smerd's logical name and its environment
	// suffix. It reproduces, byte for byte, the convention the pre-environments
	// pipeliner used (internal/pipelines/do_smerd_launch.go's
	// `req.Name = req.GetName() + "_" + p.suffix`).
	nameSuffixSeparator = "_"
)

// labelBasedRuntime is tier 1, the default: every environment on the node
// shares one Docker daemon and is kept apart by
//
//   - the labels.SuffixLabel stamped on each container (what ListSmerds and
//     friends filter on), and
//   - the actual Docker container NAME, which carries the same suffix so two
//     environments deploying the same logical smerd name don't collide on the
//     daemon's globally-unique container namespace.
//
// An empty suffix - today's default/PROD on a node that never configured
// ContainerSuffix - means "unsuffixed", preserving pre-multi-environment
// container naming exactly.
type labelBasedRuntime struct {
	commonRuntime

	suffix string
	// bakedLabels are the node's configured CustomLabels, in "name=value" (or
	// bare "name") form, stamped onto every container Velez creates - same as
	// docker.Docker's own bakedLabels.
	bakedLabels []string
}

// ContainerCreate reimplements docker.Docker.ContainerCreate's label stamping
// and conflict handling, and adds the name-conflict resolution that stamping
// alone never provided.
func (r *labelBasedRuntime) ContainerCreate(
	ctx context.Context,
	req ContainerCreateRequest,
) (container.CreateResponse, error) {
	if req.Config == nil || req.Config.Config == nil {
		return container.CreateResponse{}, rerrors.New("container config is required")
	}

	config := req.Config.Config

	if config.Labels == nil {
		config.Labels = map[string]string{}
	}

	config.Labels[labels.CreatedWithVelezLabel] = labelValueTrue
	// The suffix is the resolved environment's, decided when this runtime was
	// handed out. Empty means "the node's unscoped/default environment",
	// preserving pre-multi-environment behavior.
	config.Labels[labels.SuffixLabel] = r.suffix

	for _, label := range r.bakedLabels {
		before, after, ok := strings.Cut(label, "=")
		name := label

		var val string

		if ok {
			val = after
			name = before
		}

		config.Labels[name] = val
	}

	var hostConfig *container.HostConfig

	if req.HostConfig != nil {
		hostConfig = req.HostConfig.HostConfig
	}

	var networkingConfig *network.NetworkingConfig

	if req.NetworkingConfig != nil {
		networkingConfig = req.NetworkingConfig.NetworkingConfig
	}

	createResponse, err := r.cli.ContainerCreate(
		ctx, config, hostConfig, networkingConfig, req.Platform, r.containerName(req.ContainerName))
	if err != nil {
		if errdefs.IsConflict(err) {
			return container.CreateResponse{}, rerrors.Wrap(docker.HandleConflictMessage(err))
		}

		return container.CreateResponse{}, rerrors.Wrap(err, "error during container creation via docker api")
	}

	return createResponse, nil
}

// ListContainers scopes results to r.suffix - the environment this runtime
// instance was resolved for. The label filter is now unconditional: every
// container Velez creates (through this runtime or docker.Docker directly)
// always gets labels.SuffixLabel stamped, even when the value is "" - so an
// empty r.suffix is a real, meaningful filter value ("this environment's
// containers, whose suffix happens to be empty"), not "skip filtering and
// match everything." See docs/container_runtimes/interface_design.md.
//
// Names in each returned container.Summary are rewritten from the real
// Docker name back to the virtual/logical name (see virtualName) before
// returning - callers of ContainerRuntime never see a suffixed name.
func (r *labelBasedRuntime) ListContainers(
	ctx context.Context,
	req *velez_api.ListSmerds_Request,
) ([]container.Summary, error) {
	if req.Label == nil {
		req.Label = map[string]string{}
	}

	req.Label[labels.SuffixLabel] = r.suffix

	list, err := dockerutils.ListContainers(ctx, r.cli, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	for i := range list {
		for j, name := range list[i].Names {
			list[i].Names[j] = r.virtualName(name)
		}
	}

	return list, nil
}

// Remove deletes a single container identified by uuid or logical/Docker
// name, strictly scoped to this runtime's environment (r.suffix).
//
// Both known bugs documented on ContainerRuntime.Remove are now fixed:
//
//  1. Bare-name resolution: identifier is resolved via resolveOwnedContainer,
//     which tries the suffixed logical-name form (r.containerName(identifier))
//     before falling back to identifier as given - so a bare name in a
//     suffixed environment actually matches its real Docker container instead
//     of silently no-op'ing.
//  2. Cross-environment collision: whichever identifier form resolves to a
//     real container, its labels.SuffixLabel is compared EXACTLY against
//     r.suffix before removal proceeds - including when r.suffix is "" (empty
//     is a real value to match, not a wildcard, consistent with
//     ListContainers above). A container found under a different suffix is
//     treated as "not found in this environment": Remove returns nil without
//     touching it, the same idempotent-success semantics as a genuinely
//     absent container (internal/jobs/drop_smerd.go's dropContainerJob relies
//     on Remove never erroring for an already-gone container).
func (r *labelBasedRuntime) Remove(ctx context.Context, identifier string) error {
	resolvedID, found, err := r.resolveOwnedContainer(ctx, identifier)
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	removeOpts := container.RemoveOptions{
		Force: true,
	}

	err = r.cli.ContainerRemove(ctx, resolvedID, removeOpts)
	if err != nil {
		if strings.Contains(err.Error(), docker.NoSuchContainerError) {
			return nil
		}

		return rerrors.Wrap(err, "error removing container")
	}

	return nil
}

// Rename renames a container identified by uuid or logical/Docker name to
// newName, using the same resolveOwnedContainer resolution/ownership check as
// Remove: a bare logical name is resolved via the suffixed form first, and a
// container found under a different environment's suffix (or not found at
// all) is treated as "nothing to rename" - idempotent success, not an error.
// newName is itself passed through containerName so the renamed container
// keeps carrying this environment's suffix.
func (r *labelBasedRuntime) Rename(ctx context.Context, identifier, newName string) error {
	resolvedID, found, err := r.resolveOwnedContainer(ctx, identifier)
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	err = r.cli.ContainerRename(ctx, resolvedID, r.containerName(newName))
	if err != nil {
		if strings.Contains(err.Error(), docker.NoSuchContainerError) {
			return nil
		}

		return rerrors.Wrap(err, "error renaming container")
	}

	return nil
}

// IsContainerRunning reports whether the container identified by uuid or
// logical/Docker name is running, using the same resolveOwnedContainer
// resolution/ownership check as Remove/Rename: a container that doesn't
// exist under either identifier form, or that belongs to a different
// environment's suffix, is reported as (false, false, nil) - not an error.
// Reuses the InspectResponse resolveOwnedContainer already fetched to
// determine ownership, rather than inspecting a second time by ID - the fake
// Docker API in unit tests (and potentially a real one) only knows the
// identifier forms actually passed in, not necessarily the resolved ID as a
// distinct lookup key.
func (r *labelBasedRuntime) IsContainerRunning(ctx context.Context, identifier string) (bool, bool, error) {
	info, found, err := r.resolveOwnedContainerInfo(ctx, identifier)
	if err != nil {
		return false, false, err
	}

	if !found {
		return false, false, nil
	}

	return info.State != nil && info.State.Running, true, nil
}

// Inspect returns the container identified by uuid or logical/Docker name,
// using the same resolveOwnedContainerInfo resolution/ownership check as
// Remove/Rename/IsContainerRunning: a container that doesn't exist under
// either identifier form, or that belongs to a different environment's
// suffix, is reported as (zero value, false, nil) - not an error. The
// returned InspectResponse's Name is rewritten from the real Docker name back
// to the virtual/logical name (see virtualName), mirroring how
// ListContainers rewrites each container.Summary's Names - callers of
// ContainerRuntime never see a suffixed name.
func (r *labelBasedRuntime) Inspect(ctx context.Context, identifier string) (container.InspectResponse, bool, error) {
	info, found, err := r.resolveOwnedContainerInfo(ctx, identifier)
	if err != nil {
		return container.InspectResponse{}, false, err
	}

	if !found {
		return container.InspectResponse{}, false, nil
	}

	info.Name = r.virtualName(info.Name)

	return info, true, nil
}

// Stop stops a container identified by uuid or logical/Docker name, using the
// same resolveOwnedContainer resolution/ownership check as Remove/Rename: a
// container that doesn't exist under either identifier form, or that belongs
// to a different environment's suffix, is treated as "nothing to stop" -
// idempotent success, not an error - the same convention Remove/Rename
// established for this codebase.
func (r *labelBasedRuntime) Stop(ctx context.Context, identifier string) error {
	resolvedID, found, err := r.resolveOwnedContainer(ctx, identifier)
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	stopOpts := container.StopOptions{}

	err = r.cli.ContainerStop(ctx, resolvedID, stopOpts)
	if err != nil {
		if strings.Contains(err.Error(), docker.NoSuchContainerError) {
			return nil
		}

		return rerrors.Wrap(err, "error stopping container")
	}

	return nil
}

// Restart restarts a container identified by uuid or logical/Docker name,
// using the same resolveOwnedContainer resolution/ownership check as
// Stop/Remove/Rename: a container that doesn't exist under either identifier
// form, or that belongs to a different environment's suffix, is treated as
// "nothing to restart" - idempotent success, not an error.
func (r *labelBasedRuntime) Restart(ctx context.Context, identifier string) error {
	resolvedID, found, err := r.resolveOwnedContainer(ctx, identifier)
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	restartOpts := container.StopOptions{}

	err = r.cli.ContainerRestart(ctx, resolvedID, restartOpts)
	if err != nil {
		if strings.Contains(err.Error(), docker.NoSuchContainerError) {
			return nil
		}

		return rerrors.Wrap(err, "error restarting container")
	}

	return nil
}

// Stats returns resource-usage stats for a container identified by uuid or
// logical/Docker name, using the same resolveOwnedContainerInfo
// resolution/ownership check as Remove/Rename/IsContainerRunning/Inspect.
// Unlike those, a container that doesn't exist under either identifier form,
// or that belongs to a different environment's suffix, is a real error here
// - there is no sensible zero-value "success" for stats of a container that
// isn't there.
func (r *labelBasedRuntime) Stats(ctx context.Context, identifier string) (domain.ContainerStats, error) {
	info, found, err := r.resolveOwnedContainerInfo(ctx, identifier)
	if err != nil {
		return domain.ContainerStats{}, err
	}

	if !found {
		return domain.ContainerStats{}, rerrors.New("container %q not found in this environment", identifier)
	}

	stats, err := dockerutils.Stats(ctx, r.cli, info.ID)
	if err != nil {
		return domain.ContainerStats{}, rerrors.Wrap(err, "error getting container stats")
	}

	return stats, nil
}

// Exec runs cfg inside the container identified by uuid or logical/Docker
// name, using the same resolveOwnedContainer resolution/ownership check as
// Stop/Restart/Remove/Rename. Unlike those - which treat a missing/foreign
// container as idempotent success - and like Stats, there is no sensible
// zero-value "success" for exec output against a container that isn't
// there, so a container that doesn't exist under either identifier form, or
// belongs to a different environment's suffix, is a real error here.
//
// Reimplements docker.Docker.Exec's ContainerExecCreate/ContainerExecAttach
// sequence directly against r.cli, rather than delegating to docker.Docker -
// the same choice Stop/Restart/Remove/Rename already made (they call r.cli
// directly instead of going through docker.Docker's corresponding method).
func (r *labelBasedRuntime) Exec(
	ctx context.Context,
	containerID string,
	cfg container.ExecOptions,
) ([]byte, error) {
	resolvedID, found, err := r.resolveOwnedContainer(ctx, containerID)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, rerrors.New("container %q not found in this environment", containerID)
	}

	execResp, err := r.cli.ContainerExecCreate(ctx, resolvedID, cfg)
	if err != nil {
		return nil, rerrors.Wrap(err, "error calling exec create on container via docker api")
	}

	if !cfg.AttachStdout && !cfg.AttachStderr {
		return nil, nil
	}

	attachResp, err := r.cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, rerrors.Wrap(err, "error calling exec attach on container via docker api")
	}
	defer attachResp.Close()

	dataOut := bytes.NewBuffer(nil)

	_, err = io.Copy(dataOut, attachResp.Reader)
	if err != nil {
		return nil, rerrors.Wrap(err, "error during copying bytes from attached to container exec")
	}

	return asciiSymbolsOnly(dataOut.Bytes()), nil
}

// CreateNetwork ensures a bridge network exists for the LOGICAL name given,
// suffixed to this runtime's environment (see networkName) - creating a
// dedicated network per environment instead of the single "verv" network
// every environment used to share (docs/container_runtimes/interface_design.md).
// Idempotent: dockerutils.CreateNetwork itself no-ops if the (suffixed)
// network already exists.
func (r *labelBasedRuntime) CreateNetwork(ctx context.Context, name string) error {
	err := dockerutils.CreateNetwork(ctx, r.cli, r.networkName(name))
	if err != nil {
		return rerrors.Wrap(err, "error creating network")
	}

	return nil
}

// ConnectToNetwork connects a container to a network, both scoped to this
// runtime's environment: the container identifier is resolved via
// resolveOwnedContainer (same ownership check as Stop/Restart/Remove/Rename),
// and the network name is suffixed via networkName - mirroring CreateNetwork
// - before either is handed to dockerutils. Unlike Stop/Restart/Remove/Rename
// - which treat a missing/foreign container as idempotent success - and like
// Exec/Stats, a container not found under either identifier form, or
// belonging to a different environment, is a real error here: there's no
// sensible "connected" outcome for a container that isn't there.
func (r *labelBasedRuntime) ConnectToNetwork(ctx context.Context, req ConnectToNetworkRequest) error {
	resolvedID, found, err := r.resolveOwnedContainer(ctx, req.ContainerID)
	if err != nil {
		return err
	}

	if !found {
		return rerrors.New("container %q not found in this environment", req.ContainerID)
	}

	connectReq := dockerutils.ConnectToNetworkRequest{
		NetworkName: r.networkName(req.NetworkName),
		ContId:      resolvedID,
		Aliases:     req.Aliases,
	}

	err = dockerutils.ConnectToNetwork(ctx, r.cli, connectReq)
	if err != nil {
		return rerrors.Wrap(err, "error connecting container to network")
	}

	return nil
}

// DisconnectFromNetworks disconnects a container from each named network,
// both scoped to this runtime's environment - same resolution/ownership
// check as ConnectToNetwork, and the same networkName suffixing applied to
// every entry in networks. Same error-not-idempotent-success reasoning as
// ConnectToNetwork for a container not found under either identifier form.
func (r *labelBasedRuntime) DisconnectFromNetworks(ctx context.Context, containerID string, networks []string) error {
	resolvedID, found, err := r.resolveOwnedContainer(ctx, containerID)
	if err != nil {
		return err
	}

	if !found {
		return rerrors.New("container %q not found in this environment", containerID)
	}

	for _, n := range networks {
		err = r.cli.NetworkDisconnect(ctx, r.networkName(n), resolvedID, false)
		if err != nil {
			if strings.Contains(err.Error(), docker.NoSuchContainerError) {
				continue
			}

			return rerrors.Wrap(err, "error disconnecting container from network %q", n)
		}
	}

	return nil
}

// asciiSymbolsOnly mirrors docker.Docker's private helper of the same name -
// duplicated rather than exported/shared since Exec's whole implementation is
// intentionally a local reimplementation, not a delegation (see Exec's doc
// comment).
func asciiSymbolsOnly(in []byte) []byte {
	cleanBuff := bytes.NewBuffer(nil)

	for _, b := range in {
		if b >= 32 && b <= 127 || b == '\n' {
			cleanBuff.WriteByte(b)
		}
	}

	return cleanBuff.Bytes()
}

// resolveOwnedContainer inspects identifier - trying the suffixed logical
// name first (r.containerName(identifier)), then identifier as given (covers
// real Docker UUIDs, which must never be suffix-mangled, and any
// already-correct raw name) - and returns the real container ID if and only
// if a container was found AND its labels.SuffixLabel exactly equals
// r.suffix.
//
// found is false, with no error, in two cases that both mean "nothing for
// Remove to do": no container exists under either identifier form, or a
// container was found but belongs to a different environment (a real,
// suffix-labeled container that just isn't r's to remove). Only an
// unexpected inspect failure (Docker unreachable, etc.) is surfaced as err.
func (r *labelBasedRuntime) resolveOwnedContainer(
	ctx context.Context,
	identifier string,
) (id string, found bool, err error) {
	info, found, err := r.resolveOwnedContainerInfo(ctx, identifier)
	if err != nil || !found {
		return "", found, err
	}

	return info.ID, true, nil
}

// resolveOwnedContainerInfo is resolveOwnedContainer's implementation,
// additionally returning the full InspectResponse so callers that need more
// than the ID (IsContainerRunning's State) don't have to inspect the
// container a second time under a different identifier form.
func (r *labelBasedRuntime) resolveOwnedContainerInfo(
	ctx context.Context,
	identifier string,
) (container.InspectResponse, bool, error) {
	for _, candidate := range r.candidateNames(identifier) {
		info, inspectErr := r.cli.ContainerInspect(ctx, candidate)
		if inspectErr != nil {
			if strings.Contains(inspectErr.Error(), docker.NoSuchContainerError) {
				continue
			}

			return container.InspectResponse{}, false, rerrors.Wrap(inspectErr, "error inspecting container")
		}

		var suffix string

		if info.Config != nil {
			suffix = info.Config.Labels[labels.SuffixLabel]
		}

		if suffix != r.suffix {
			return container.InspectResponse{}, false, nil
		}

		return info, true, nil
	}

	return container.InspectResponse{}, false, nil
}

// candidateNames returns the identifier forms resolveOwnedContainer should
// try, in order: the suffixed logical name first, then identifier itself.
// When r.suffix is empty, containerName(identifier) already equals
// identifier, so the second attempt would be redundant - it's omitted.
func (r *labelBasedRuntime) candidateNames(identifier string) []string {
	suffixed := r.containerName(identifier)
	if suffixed == identifier {
		return []string{identifier}
	}

	return []string{suffixed, identifier}
}

// containerName resolves the logical smerd name into the actual Docker
// container name for this environment: the bare name when the suffix is empty,
// "<name>_<suffix>" otherwise.
func (r *labelBasedRuntime) containerName(name string) string {
	if r.suffix == "" {
		return name
	}

	return name + nameSuffixSeparator + r.suffix
}

// networkName resolves a logical network name into the actual Docker network
// name for this environment - byte-for-byte the same suffixing rule
// containerName applies to container names (see nameSuffixSeparator): the
// bare name when the suffix is empty, "<name>_<suffix>" otherwise. Kept as
// its own named method (rather than callers using containerName directly)
// so a future divergence between container- and network-naming rules doesn't
// require re-auditing every call site.
func (r *labelBasedRuntime) networkName(name string) string {
	return r.containerName(name)
}

// virtualName reverses containerName: strips this runtime's suffix from a
// real Docker name, recovering the virtual/logical name ContainerCreate was
// called with. Delegates to the package-level StripEnvironmentSuffix so the
// "<name>_<suffix>" convention is defined in exactly one place - see that
// function's doc comment for why it's exported even though this is its only
// caller today.
func (r *labelBasedRuntime) virtualName(dockerName string) string {
	return StripEnvironmentSuffix(dockerName, r.suffix)
}

// StripEnvironmentSuffix reverses the "<name>_<suffix>" naming convention
// labelBasedRuntime.containerName applies (byte for byte - see
// nameSuffixSeparator): given a real Docker name and the suffix it was
// created with, it returns the virtual/logical name. An empty suffix or a
// name that doesn't end in "_<suffix>" is returned unchanged.
//
// Exported (rather than folded into virtualName) because, historically, not
// every reader of a Docker container name went through a resolved
// ContainerRuntime - container_manager.InspectSmerd used to call
// client.APIClient.ContainerInspect directly and recover the suffix from the
// container's own labels.SuffixLabel. InspectSmerd itself has since been
// rewired onto ContainerRuntime.Inspect (which calls this via virtualName),
// so virtualName is this function's only caller today - kept exported since
// other packages/tests may still reference it directly (see
// docs/container_runtimes/roadmap.md).
func StripEnvironmentSuffix(dockerName, suffix string) string {
	if suffix == "" {
		return dockerName
	}

	return strings.TrimSuffix(dockerName, nameSuffixSeparator+suffix)
}
