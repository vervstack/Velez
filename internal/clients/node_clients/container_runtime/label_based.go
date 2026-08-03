package container_runtime

import (
	"context"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
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
	for _, candidate := range r.candidateNames(identifier) {
		info, inspectErr := r.cli.ContainerInspect(ctx, candidate)
		if inspectErr != nil {
			if strings.Contains(inspectErr.Error(), docker.NoSuchContainerError) {
				continue
			}

			return "", false, rerrors.Wrap(inspectErr, "error inspecting container")
		}

		var suffix string

		if info.Config != nil {
			suffix = info.Config.Labels[labels.SuffixLabel]
		}

		if suffix != r.suffix {
			return "", false, nil
		}

		return info.ID, true, nil
	}

	return "", false, nil
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

// virtualName reverses containerName: strips this runtime's suffix from a
// real Docker name, recovering the virtual/logical name ContainerCreate was
// called with. Delegates to the package-level StripEnvironmentSuffix so the
// "<name>_<suffix>" convention is defined in exactly one place, shared with
// callers outside this package that need the same reversal from a suffix
// read off a container's own label rather than from a resolved runtime (see
// container_manager.InspectSmerd).
func (r *labelBasedRuntime) virtualName(dockerName string) string {
	return StripEnvironmentSuffix(dockerName, r.suffix)
}

// StripEnvironmentSuffix reverses the "<name>_<suffix>" naming convention
// labelBasedRuntime.containerName applies (byte for byte - see
// nameSuffixSeparator): given a real Docker name and the suffix it was
// created with, it returns the virtual/logical name. An empty suffix or a
// name that doesn't end in "_<suffix>" is returned unchanged.
//
// Exported because not every reader of a Docker container name goes through
// a resolved ContainerRuntime yet - container_manager.InspectSmerd (Phase 1
// scope, see docs/container_runtimes/roadmap.md) still calls
// client.APIClient.ContainerInspect directly and recovers the suffix from
// the container's own labels.SuffixLabel. Exporting this keeps the suffix
// convention itself defined once, even though it's invoked from both places.
func StripEnvironmentSuffix(dockerName, suffix string) string {
	if suffix == "" {
		return dockerName
	}

	return strings.TrimSuffix(dockerName, nameSuffixSeparator+suffix)
}
