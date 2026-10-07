package container_runtime

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/clients/node_clients/runtime_policy"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	labelValueTrue = "true"
)

var _ ContainerRuntime = (*dockerRuntime)(nil)

// dockerRuntime is the single shared implementation of every
// environment-scoped ContainerRuntime method, generic and backend-agnostic
// over HOW names/labels are scoped to an environment - that policy is
// delegated to whichever nameResolver it holds. Today's only resolver,
// labelSuffixResolver, implements tier 1 (every environment on the node
// shares one Docker daemon, kept apart by suffixed names/labels); a future
// tier 2 (a Docker daemon dedicated to one environment) plugs in via
// directResolver without duplicating any of these 12 methods.
type dockerRuntime struct {
	commonRuntime

	resolver nameResolver
	// bakedLabels are the node's configured CustomLabels, in "name=value" (or
	// bare "name") form, stamped onto every container Velez creates - same as
	// docker.Docker's own bakedLabels.
	bakedLabels []string
	// settings is nil when node-wide settings are disabled.
	settings SettingsProvider
}

// newLabelBasedRuntime builds a dockerRuntime backed by labelSuffixResolver -
// tier 1, the default: one shared Docker daemon, environments kept apart by
// suffix labels/names.
func newLabelBasedRuntime(cli client.APIClient, suffix string, bakedLabels []string) *dockerRuntime {
	common := commonRuntime{
		cli: cli,
	}

	resolver := &labelSuffixResolver{
		suffix: suffix,
	}

	return &dockerRuntime{
		commonRuntime: common,
		resolver:      resolver,
		bakedLabels:   bakedLabels,
	}
}

// newDirectRuntime builds a dockerRuntime backed by directResolver - tier 2,
// a Docker daemon dedicated to a single environment, needing no name/label
// scoping at all. Wired into production by resolver.go's dedicated branch.
func newDirectRuntime(cli client.APIClient, bakedLabels []string) *dockerRuntime {
	common := commonRuntime{
		cli: cli,
	}

	resolver := &directResolver{}

	return &dockerRuntime{
		commonRuntime: common,
		resolver:      resolver,
		bakedLabels:   bakedLabels,
	}
}

// ContainerCreate reimplements docker.Docker.ContainerCreate's label stamping
// and conflict handling, and adds the name-conflict resolution that stamping
// alone never provided.
func (r *dockerRuntime) ContainerCreate(
	ctx context.Context,
	req ContainerCreateRequest,
) (container.CreateResponse, error) {
	config := req.Config.Config

	if config.Labels == nil {
		config.Labels = map[string]string{}
	}

	config.Labels[labels.CreatedWithVelezLabel] = labelValueTrue
	// The resolver's scoping is the resolved environment's, decided when this
	// runtime was handed out.
	r.resolver.StampLabels(config.Labels)

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

	err := runtime_policy.Resolve(ctx, r.settings, hostConfig, config.Image)
	if err != nil {
		return container.CreateResponse{}, rerrors.Wrap(err, "error resolving runtime policy")
	}

	createResponse, err := r.cli.ContainerCreate(
		ctx, config, hostConfig, networkingConfig, req.Platform, r.resolver.ContainerName(req.ContainerName))
	if err != nil {
		if errdefs.IsConflict(err) {
			return container.CreateResponse{}, rerrors.Wrap(docker.HandleConflictMessage(err))
		}

		return container.CreateResponse{}, rerrors.Wrap(err, "error during container creation via docker api")
	}

	return createResponse, nil
}

// ListContainers scopes results to this runtime's environment - the label
// filter is now unconditional: every container Velez creates (through this
// runtime or docker.Docker directly) always gets labels.SuffixLabel stamped,
// even when the value is "" - so an empty suffix is a real, meaningful
// filter value ("this environment's containers, whose suffix happens to be
// empty"), not "skip filtering and match everything." See
// docs/container_runtimes/interface_design.md.
//
// Names in each returned container.Summary are rewritten from the real
// Docker name back to the virtual/logical name (see
// nameResolver.VirtualContainerName) before returning - callers of
// ContainerRuntime never see a suffixed name.
func (r *dockerRuntime) ListContainers(
	ctx context.Context,
	req *velez_api.ListSmerds_Request,
) ([]container.Summary, error) {
	if req.Label == nil {
		req.Label = map[string]string{}
	}

	r.resolver.ListFilterLabels(req.GetLabel())

	list, err := dockerutils.ListContainers(ctx, r.cli, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	for i := range list {
		for j, name := range list[i].Names {
			list[i].Names[j] = r.resolver.VirtualContainerName(name)
		}
	}

	return list, nil
}

// Remove deletes a single container identified by uuid or logical/Docker
// name, strictly scoped to this runtime's environment.
//
// Both known bugs documented on ContainerRuntime.Remove are now fixed:
//
//  1. Bare-name resolution: identifier is resolved via resolveOwnedContainer,
//     which tries the suffixed logical-name form
//     (r.resolver.ContainerName(identifier)) before falling back to
//     identifier as given - so a bare name in a suffixed environment
//     actually matches its real Docker container instead of silently
//     no-op'ing.
//  2. Cross-environment collision: whichever identifier form resolves to a
//     real container, its ownership is checked via r.resolver.Owns before
//     removal proceeds. A container found under a different environment is
//     treated as "not found in this environment": Remove returns nil without
//     touching it, the same idempotent-success semantics as a genuinely
//     absent container (internal/jobs/drop_smerd.go's dropContainerJob relies
//     on Remove never erroring for an already-gone container).
func (r *dockerRuntime) Remove(ctx context.Context, identifier string) error {
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
// container found under a different environment is treated as "nothing to
// rename" - idempotent success, not an error. newName is itself passed
// through the resolver's ContainerName so the renamed container keeps
// carrying this environment's scoping.
func (r *dockerRuntime) Rename(ctx context.Context, identifier, newName string) error {
	resolvedID, found, err := r.resolveOwnedContainer(ctx, identifier)
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	err = r.cli.ContainerRename(ctx, resolvedID, r.resolver.ContainerName(newName))
	if err != nil {
		if strings.Contains(err.Error(), docker.NoSuchContainerError) {
			return nil
		}

		return rerrors.Wrap(err, "error renaming container")
	}

	return nil
}

// IsContainerRunning reports whether the container identified by uuid or
// logical/Docker name is running, using the same resolveOwnedContainerInfo
// resolution/ownership check as Remove/Rename: a container that doesn't
// exist under either identifier form, or that belongs to a different
// environment, is reported as (false, false, nil) - not an error. Reuses the
// InspectResponse resolveOwnedContainerInfo already fetched to determine
// ownership, rather than inspecting a second time by ID - the fake Docker API
// in unit tests (and potentially a real one) only knows the identifier forms
// actually passed in, not necessarily the resolved ID as a distinct lookup
// key.
func (r *dockerRuntime) IsContainerRunning(ctx context.Context, identifier string) (bool, bool, error) {
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
// either identifier form, or that belongs to a different environment, is
// reported as (zero value, false, nil) - not an error. The returned
// InspectResponse's Name is rewritten from the real Docker name back to the
// virtual/logical name (see nameResolver.VirtualContainerName), mirroring how
// ListContainers rewrites each container.Summary's Names - callers of
// ContainerRuntime never see a suffixed name.
func (r *dockerRuntime) Inspect(ctx context.Context, identifier string) (container.InspectResponse, bool, error) {
	info, found, err := r.resolveOwnedContainerInfo(ctx, identifier)
	if err != nil {
		return container.InspectResponse{}, false, err
	}

	if !found {
		return container.InspectResponse{}, false, nil
	}

	info.Name = r.resolver.VirtualContainerName(info.Name)

	return info, true, nil
}

// Stop stops a container identified by uuid or logical/Docker name, using the
// same resolveOwnedContainer resolution/ownership check as Remove/Rename: a
// container that doesn't exist under either identifier form, or that belongs
// to a different environment, is treated as "nothing to stop" - idempotent
// success, not an error - the same convention Remove/Rename established for
// this codebase.
func (r *dockerRuntime) Stop(ctx context.Context, identifier string) error {
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
// form, or that belongs to a different environment, is treated as "nothing
// to restart" - idempotent success, not an error.
func (r *dockerRuntime) Restart(ctx context.Context, identifier string) error {
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
// or that belongs to a different environment, is a real error here - there
// is no sensible zero-value "success" for stats of a container that isn't
// there.
func (r *dockerRuntime) Stats(ctx context.Context, identifier string) (domain.ContainerStats, error) {
	info, found, err := r.resolveOwnedContainerInfo(ctx, identifier)
	if err != nil {
		return domain.ContainerStats{}, err
	}

	if !found {
		return domain.ContainerStats{}, user_errors.ErrNoSuchContainer
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
// belongs to a different environment, is a real error here.
//
// Reimplements docker.Docker.Exec's ContainerExecCreate/ContainerExecAttach
// sequence directly against r.cli, rather than delegating to docker.Docker -
// the same choice Stop/Restart/Remove/Rename already made (they call r.cli
// directly instead of going through docker.Docker's corresponding method).
// exitCode is only meaningful once the exec has actually run, which only
// happens on the attached path below - the unattached early return reports
// exitCode 0 unconditionally, since ContainerExecCreate alone never starts
// the process.
func (r *dockerRuntime) Exec(
	ctx context.Context,
	containerID string,
	cfg container.ExecOptions,
) ([]byte, int, error) {
	resolvedID, found, err := r.resolveOwnedContainer(ctx, containerID)
	if err != nil {
		return nil, 0, err
	}

	if !found {
		return nil, 0, user_errors.ErrNoSuchContainer
	}

	execResp, err := r.cli.ContainerExecCreate(ctx, resolvedID, cfg)
	if err != nil {
		return nil, 0, rerrors.Wrap(err, "error calling exec create on container via docker api")
	}

	if !cfg.AttachStdout && !cfg.AttachStderr {
		return nil, 0, nil
	}

	attachResp, err := r.cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, 0, rerrors.Wrap(err, "error calling exec attach on container via docker api")
	}
	defer attachResp.Close()

	dataOut := bytes.NewBuffer(nil)

	_, err = io.Copy(dataOut, attachResp.Reader)
	if err != nil {
		return nil, 0, rerrors.Wrap(err, "error during copying bytes from attached to container exec")
	}

	inspectResp, err := r.cli.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return nil, 0, rerrors.Wrap(err, "error inspecting container exec result")
	}

	return asciiSymbolsOnly(dataOut.Bytes()), inspectResp.ExitCode, nil
}

// CopyFromContainer reads the single file at filePath out of the tar stream
// the Docker API returns for it.
func (r *dockerRuntime) CopyFromContainer(ctx context.Context, containerID, filePath string) ([]byte, error) {
	resolvedID, found, err := r.resolveOwnedContainer(ctx, containerID)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, user_errors.ErrNoSuchContainer
	}

	archive, _, err := r.cli.CopyFromContainer(ctx, resolvedID, filePath)
	if err != nil {
		return nil, rerrors.Wrap(err, "error copying file from container via docker api")
	}
	defer common.CloseWithLog(archive.Close, "container file archive")

	reader := tar.NewReader(archive)

	for {
		header, nextErr := reader.Next()
		if errors.Is(nextErr, io.EOF) {
			return nil, rerrors.Wrap(user_errors.ErrContainerFileNotFound)
		}

		if nextErr != nil {
			return nil, rerrors.Wrap(nextErr, "error reading tar stream from container")
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		content, readErr := io.ReadAll(reader)
		if readErr != nil {
			return nil, rerrors.Wrap(readErr, "error reading file content from container tar stream")
		}

		return content, nil
	}
}

// CopyToContainer wraps content into a one-entry tar and extracts it into
// filePath's directory, because the Docker API only accepts a directory
// destination plus a tar stream.
func (r *dockerRuntime) CopyToContainer(
	ctx context.Context,
	containerID, filePath string,
	content []byte,
	mode fs.FileMode,
) error {
	resolvedID, found, err := r.resolveOwnedContainer(ctx, containerID)
	if err != nil {
		return err
	}

	if !found {
		return user_errors.ErrNoSuchContainer
	}

	archive := bytes.NewBuffer(nil)
	writer := tar.NewWriter(archive)

	header := &tar.Header{
		Name:     path.Base(filePath),
		Typeflag: tar.TypeReg,
		Mode:     int64(mode.Perm()),
		Size:     int64(len(content)),
	}

	err = writer.WriteHeader(header)
	if err != nil {
		return rerrors.Wrap(err, "error writing tar header")
	}

	_, err = writer.Write(content)
	if err != nil {
		return rerrors.Wrap(err, "error writing tar content")
	}

	err = writer.Close()
	if err != nil {
		return rerrors.Wrap(err, "error closing tar writer")
	}

	err = r.cli.CopyToContainer(ctx, resolvedID, path.Dir(filePath), archive, container.CopyToContainerOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error copying file to container via docker api")
	}

	return nil
}

// CreateNetwork ensures a bridge network exists for the LOGICAL name given,
// scoped to this runtime's environment (see nameResolver.NetworkName) -
// creating a dedicated network per environment instead of the single "verv"
// network every environment used to share
// (docs/container_runtimes/interface_design.md). Idempotent:
// dockerutils.CreateNetwork itself no-ops if the (scoped) network already
// exists.
func (r *dockerRuntime) CreateNetwork(ctx context.Context, name string) error {
	err := dockerutils.CreateNetwork(ctx, r.cli, r.resolver.NetworkName(name))
	if err != nil {
		return rerrors.Wrap(err, "error creating network")
	}

	return nil
}

// ConnectToNetwork connects a container to a network, both scoped to this
// runtime's environment: the container identifier is resolved via
// resolveOwnedContainer (same ownership check as Stop/Restart/Remove/Rename),
// and the network name is scoped via r.resolver.NetworkName - mirroring
// CreateNetwork - before either is handed to dockerutils. Unlike
// Stop/Restart/Remove/Rename - which treat a missing/foreign container as
// idempotent success - and like Exec/Stats, a container not found under
// either identifier form, or belonging to a different environment, is a real
// error here: there's no sensible "connected" outcome for a container that
// isn't there.
func (r *dockerRuntime) ConnectToNetwork(ctx context.Context, req ConnectToNetworkRequest) error {
	resolvedID, found, err := r.resolveOwnedContainer(ctx, req.ContainerID)
	if err != nil {
		return err
	}

	if !found {
		return user_errors.ErrNoSuchContainer
	}

	connectReq := dockerutils.ConnectToNetworkRequest{
		NetworkName: r.resolver.NetworkName(req.NetworkName),
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
// check as ConnectToNetwork, and the same r.resolver.NetworkName scoping
// applied to every entry in networks. Same error-not-idempotent-success
// reasoning as ConnectToNetwork for a container not found under either
// identifier form.
func (r *dockerRuntime) DisconnectFromNetworks(ctx context.Context, containerID string, networks []string) error {
	resolvedID, found, err := r.resolveOwnedContainer(ctx, containerID)
	if err != nil {
		return err
	}

	if !found {
		return user_errors.ErrNoSuchContainer
	}

	for _, n := range networks {
		err = r.cli.NetworkDisconnect(ctx, r.resolver.NetworkName(n), resolvedID, false)
		if err != nil {
			isAbsent := strings.Contains(err.Error(), docker.NoSuchContainerError) ||
				strings.Contains(err.Error(), docker.NotConnectedToNetworkError)
			if isAbsent {
				continue
			}

			return rerrors.Wrapf(err, "error disconnecting container from network %q", n)
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

func (r *dockerRuntime) ListNetworks(ctx context.Context, isForeignIncluded bool) ([]NetworkInfo, error) {
	listOpts := network.ListOptions{}
	inspectOpts := network.InspectOptions{}

	summaries, err := r.cli.NetworkList(ctx, listOpts)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing networks")
	}

	infos := make([]NetworkInfo, 0, len(summaries))

	for _, summary := range summaries {
		state := classifyNetwork(summary, r.resolver)

		isSkipped := state == networkStateOtherEnvironment ||
			(state != networkStateManaged && !isForeignIncluded)
		if isSkipped {
			continue
		}

		inspected, inspectErr := r.cli.NetworkInspect(ctx, summary.ID, inspectOpts)
		if inspectErr != nil {
			if errdefs.IsNotFound(inspectErr) {
				continue
			}

			return nil, rerrors.Wrap(inspectErr, "error inspecting network")
		}

		info := buildNetworkInfo(inspected, state, r.resolver)
		r.fillMemberAliases(ctx, &info)

		infos = append(infos, info)
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})

	return infos, nil
}

func (r *dockerRuntime) InspectNetwork(ctx context.Context, networkId string) (NetworkInfo, error) {
	info, _, err := r.inspectScopedNetwork(ctx, networkId)
	if err != nil {
		return NetworkInfo{}, rerrors.Wrap(err)
	}

	return info, nil
}

func (r *dockerRuntime) CreateManagedNetwork(ctx context.Context, req CreateNetworkRequest) (NetworkInfo, error) {
	dockerName := r.resolver.NetworkName(req.Name)

	listOpts := network.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", dockerName)),
	}

	existing, err := r.cli.NetworkList(ctx, listOpts)
	if err != nil {
		return NetworkInfo{}, rerrors.Wrap(err, "error listing networks")
	}

	for _, item := range existing {
		if item.Name == dockerName {
			return NetworkInfo{}, rerrors.Wrap(user_errors.ErrNetworkAlreadyExists)
		}
	}

	networkLabels := map[string]string{labels.NetworkManagedLabel: labels.NetworkManagedLabelValue}
	r.resolver.StampLabels(networkLabels)

	opts := dockerutils.CreateNetworkOptions{
		Labels:       networkLabels,
		IsInternal:   req.IsInternal,
		IsIccEnabled: req.IsIccEnabled,
	}

	id, err := dockerutils.CreateNetworkWithOptions(ctx, r.cli, dockerName, opts)
	if err != nil {
		return NetworkInfo{}, rerrors.Wrap(err, "error creating network")
	}

	info, _, err := r.inspectScopedNetwork(ctx, id)
	if err != nil {
		return NetworkInfo{}, rerrors.Wrap(err)
	}

	return info, nil
}

func (r *dockerRuntime) RemoveNetwork(ctx context.Context, networkId string) error {
	info, state, err := r.inspectScopedNetwork(ctx, networkId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if state != networkStateManaged {
		return rerrors.Wrap(user_errors.ErrNetworkNotManaged)
	}

	err = r.cli.NetworkRemove(ctx, info.Id)
	if err != nil {
		return rerrors.Wrap(err, "error removing network")
	}

	return nil
}

func (r *dockerRuntime) AttachContainer(ctx context.Context, req AttachContainerRequest) error {
	resolvedId, info, err := r.resolveContainerAndNetwork(ctx, req.ContainerID, req.NetworkId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	connectReq := dockerutils.ConnectToNetworkRequest{
		NetworkName: info.DockerName,
		ContId:      resolvedId,
		Aliases:     req.Aliases,
	}

	err = dockerutils.ConnectToNetwork(ctx, r.cli, connectReq)
	if err != nil {
		return rerrors.Wrap(err, "error attaching container to network")
	}

	return nil
}

func (r *dockerRuntime) DetachContainer(ctx context.Context, containerId, networkId string) error {
	resolvedId, info, err := r.resolveContainerAndNetwork(ctx, containerId, networkId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = r.cli.NetworkDisconnect(ctx, info.Id, resolvedId, false)
	if err != nil {
		isAbsent := strings.Contains(err.Error(), docker.NoSuchContainerError) ||
			strings.Contains(err.Error(), docker.NotConnectedToNetworkError)
		if isAbsent {
			return nil
		}

		return rerrors.Wrap(err, "error detaching container from network")
	}

	return nil
}

func (r *dockerRuntime) resolveContainerAndNetwork(
	ctx context.Context,
	containerIdentifier, networkId string,
) (containerId string, info NetworkInfo, err error) {
	containerId, found, err := r.resolveOwnedContainer(ctx, containerIdentifier)
	if err != nil {
		return "", NetworkInfo{}, rerrors.Wrap(err)
	}

	if !found {
		return "", NetworkInfo{}, rerrors.Wrap(user_errors.ErrNoSuchContainer)
	}

	info, state, err := r.inspectScopedNetwork(ctx, networkId)
	if err != nil {
		return "", NetworkInfo{}, rerrors.Wrap(err)
	}

	if state == networkStateSystem {
		return "", NetworkInfo{}, rerrors.Wrap(user_errors.ErrNetworkNotManaged)
	}

	return containerId, info, nil
}

// inspectScopedNetwork inspects a network and treats one that belongs to another environment as not found.
func (r *dockerRuntime) inspectScopedNetwork(
	ctx context.Context,
	networkId string,
) (NetworkInfo, networkState, error) {
	inspectOpts := network.InspectOptions{}

	inspected, err := r.cli.NetworkInspect(ctx, networkId, inspectOpts)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return NetworkInfo{}, networkStateForeign, rerrors.Wrap(user_errors.ErrNetworkNotFound)
		}

		return NetworkInfo{}, networkStateForeign, rerrors.Wrap(err, "error inspecting network")
	}

	state := classifyNetwork(inspected, r.resolver)
	if state == networkStateOtherEnvironment {
		return NetworkInfo{}, networkStateForeign, rerrors.Wrap(user_errors.ErrNetworkNotFound)
	}

	info := buildNetworkInfo(inspected, state, r.resolver)
	r.fillMemberAliases(ctx, &info)

	return info, state, nil
}

// fillMemberAliases is best effort: a member whose container cannot be inspected keeps no aliases.
func (r *dockerRuntime) fillMemberAliases(ctx context.Context, info *NetworkInfo) {
	for i := range info.Members {
		member := &info.Members[i]

		cont, err := r.cli.ContainerInspect(ctx, member.ContainerId)
		if err != nil || cont.NetworkSettings == nil {
			continue
		}

		endpoint := cont.NetworkSettings.Networks[info.DockerName]
		if endpoint == nil {
			continue
		}

		member.Aliases = filterAliases(endpoint.Aliases, member.ContainerId, cont.Name)
	}
}

// resolveOwnedContainer inspects identifier - trying the resolver's scoped
// logical name first (r.resolver.ContainerName(identifier)), then identifier
// as given (covers real Docker UUIDs, which must never be name-mangled, and
// any already-correct raw name) - and returns the real container ID if and
// only if a container was found AND r.resolver.Owns its labels.
//
// found is false, with no error, in two cases that both mean "nothing for
// Remove to do": no container exists under either identifier form, or a
// container was found but belongs to a different environment (a real,
// labeled container that just isn't r's to remove). Only an unexpected
// inspect failure (Docker unreachable, etc.) is surfaced as err.
func (r *dockerRuntime) resolveOwnedContainer(
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
func (r *dockerRuntime) resolveOwnedContainerInfo(
	ctx context.Context,
	identifier string,
) (container.InspectResponse, bool, error) {
	for _, candidate := range r.resolver.CandidateNames(identifier) {
		info, inspectErr := r.cli.ContainerInspect(ctx, candidate)
		if inspectErr != nil {
			if strings.Contains(inspectErr.Error(), docker.NoSuchContainerError) {
				continue
			}

			return container.InspectResponse{}, false, rerrors.Wrap(inspectErr, "error inspecting container")
		}

		var containerLabels map[string]string

		if info.Config != nil {
			containerLabels = info.Config.Labels
		}

		if !r.resolver.Owns(containerLabels) {
			return container.InspectResponse{}, false, nil
		}

		return info, true, nil
	}

	return container.InspectResponse{}, false, nil
}
