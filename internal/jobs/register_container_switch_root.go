package jobs

import (
	"context"
	"maps"
	"slices"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/runtime_policy"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	stepSwitchRoot = "switch_root"

	defaultNetworkMode   = "default"
	defaultBridgeNetwork = "bridge"
	hostNetwork          = "host"
	noneNetwork          = "none"
)

// switchRootPlan names the containers of a root switch: NewRoot takes over the network namespace owned by OldRoot
// (its published ports, networks and aliases), then OldRoot and every other member rejoin it.
type switchRootPlan struct {
	OldRootId   string
	OldRootName string
	NewRootId   string
	NewRootName string
	// IsNewRootService is set when the new root is the service container.
	IsNewRootService bool
	// Members rejoin the new root, the old root first.
	Members []string
}

func newSwitchRootPlan(payload *velez_api.RegisterContainerTaskPayload) switchRootPlan {
	plan := switchRootPlan{
		OldRootId:        payload.GetPreviousRootContainerId(),
		OldRootName:      payload.GetPreviousRootContainerName(),
		NewRootId:        payload.GetRootContainerId(),
		NewRootName:      payload.GetRootContainerName(),
		IsNewRootService: payload.GetRootContainerId() == "",
	}

	if plan.IsNewRootService {
		plan.NewRootId = payload.GetContainerId()
		plan.NewRootName = payload.GetContainerName()
	}

	candidates := slices.Concat(
		[]string{plan.OldRootName},
		payload.GetSidecarContainerNames(),
		[]string{payload.GetContainerName()},
	)

	for _, name := range candidates {
		isKnown := name == plan.NewRootName || slices.Contains(plan.Members, name)
		if !isKnown {
			plan.Members = append(plan.Members, name)
		}
	}

	return plan
}

// switchRootJob makes another member of the network group the owner of the network namespace. Published host ports
// are exclusive, so the swap is not blue-green: the old root and the new root's old instance are stopped and
// renamed aside, the new root is created with the old root's ports, networks and aliases, then the old root and
// every other member are recreated joined to it. All or nothing: Rollback puts the originals back, up to the point
// the switch is committed.
type switchRootJob struct {
	runtimes container_runtime.RuntimeResolver
	payload  *velez_api.RegisterContainerTaskPayload
}

func newSwitchRootJob(
	runtimes container_runtime.RuntimeResolver, payload *velez_api.RegisterContainerTaskPayload,
) *switchRootJob {
	return &switchRootJob{runtimes: runtimes, payload: payload}
}

func (j *switchRootJob) Do(ctx context.Context) error {
	plan := newSwitchRootPlan(j.payload)

	runtime, err := j.runtimes.Runtime(ctx, j.payload.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	oldRoot, err := inspectSwitchContainer(ctx, runtime, plan.OldRootId)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting the old network root")
	}

	newRoot, err := inspectSwitchContainer(ctx, runtime, plan.NewRootId)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting the new network root")
	}

	err = checkMembersMayRejoin(ctx, runtime, plan, oldRoot)
	if err != nil {
		return rerrors.Wrap(err)
	}

	sidecars := j.sidecars()

	err = releaseOriginal(ctx, runtime, sidecars, plan.OldRootName, oldRoot.ID)
	if err != nil {
		return rerrors.Wrap(err, "error releasing the old network root")
	}

	err = releaseOriginal(ctx, runtime, sidecars, plan.NewRootName, newRoot.ID)
	if err != nil {
		return rerrors.Wrap(err, "error releasing the new network root")
	}

	newRootId, err := j.createNewRoot(ctx, runtime, plan, oldRoot, newRoot)
	if err != nil {
		return rerrors.Wrap(err, "error creating the new network root")
	}

	networkMode := container.NetworkMode(networkModeOfPrefix + newRootId)

	for _, name := range plan.Members {
		err = sidecars.recreate(ctx, runtime, name, networkMode)
		if err != nil {
			return rerrors.Wrap(err, "error rejoining "+name)
		}
	}

	service, isFound, err := runtime.Inspect(ctx, j.payload.GetContainerName())
	if err != nil {
		return rerrors.Wrap(err, "error inspecting the service container")
	}

	if !isFound {
		return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound, j.payload.GetContainerName())
	}

	j.payload.SetSwitchedServiceContainerId(service.ID)

	err = dropSwitchBackups(ctx, runtime, plan)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

// Rollback drops what Do created and puts the originals back, the old root first because the members can only
// start next to a running namespace owner. A committed switch is left as it is.
func (j *switchRootJob) Rollback(ctx context.Context) error {
	if j.payload.GetSwitchedServiceContainerId() != "" {
		return nil
	}

	plan := newSwitchRootPlan(j.payload)

	runtime, err := j.runtimes.Runtime(ctx, j.payload.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	sidecars := j.sidecars()
	group := registerSidecarGroup{payload: j.payload}
	failures := make([]error, 0)

	for _, name := range plan.Members {
		isOurs := func(containerLabels map[string]string) bool {
			return isRegisteredMemberOf(group, name, containerLabels)
		}

		_, err = sidecars.restoreAside(ctx, runtime, name, isOurs)
		if err != nil {
			failures = append(failures, rerrors.Wrap(err, "error restoring "+name))
		}
	}

	isNewRootOurs := func(containerLabels map[string]string) bool {
		return containerLabels[labels.OnboardedFromLabel] == plan.OldRootId
	}

	_, err = sidecars.restoreAside(ctx, runtime, plan.NewRootName, isNewRootOurs)
	if err != nil {
		failures = append(failures, rerrors.Wrap(err, "error restoring "+plan.NewRootName))
	}

	startOrder := make([]string, 0, len(plan.Members)+1)

	startOrder = append(startOrder, plan.OldRootName, plan.NewRootName)
	startOrder = append(startOrder, plan.Members...)

	started := make(map[string]bool, len(startOrder))

	for _, name := range startOrder {
		if started[name] {
			continue
		}

		started[name] = true

		err = runtime.Restart(ctx, name)
		if err != nil {
			failures = append(failures, rerrors.Wrap(err, "error restarting "+name))
		}
	}

	err = rerrors.Join(failures...)
	if err != nil {
		return rerrors.Wrap(err, "error rolling back the root switch")
	}

	return nil
}

func (j *switchRootJob) sidecars() *recreateSidecarsJob {
	return newRecreateSidecarsJob(j.runtimes, registerSidecarGroup{payload: j.payload})
}

func inspectSwitchContainer(
	ctx context.Context, runtime container_runtime.ContainerRuntime, id string,
) (container.InspectResponse, error) {
	info, isFound, err := runtime.InspectAny(ctx, id)
	if err != nil {
		return container.InspectResponse{}, rerrors.Wrap(err)
	}

	if !isFound || info.Config == nil || info.HostConfig == nil {
		return container.InspectResponse{}, rerrors.Wrap(user_errors.ErrRegisterContainerNotFound, id)
	}

	return info, nil
}

// checkMembersMayRejoin refuses the switch before anything is touched when the runtime policy would reject a
// member joining another container's network namespace.
func checkMembersMayRejoin(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	plan switchRootPlan,
	oldRoot container.InspectResponse,
) error {
	images := []string{oldRoot.Config.Image}

	for _, name := range plan.Members {
		if name == plan.OldRootName {
			continue
		}

		member, isFound, err := runtime.Inspect(ctx, name)
		if err != nil {
			return rerrors.Wrap(err, "error inspecting "+name)
		}

		if !isFound {
			return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound, name)
		}

		images = append(images, member.Config.Image)
	}

	probe := &container.HostConfig{NetworkMode: container.NetworkMode(networkModeOfPrefix + plan.NewRootId)}

	for _, image := range images {
		err := runtime_policy.Apply(probe, image, domain.Settings{})
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	return nil
}

// releaseOriginal stops the original container and renames it to its backup name, which frees its name and host
// ports. A container that is no longer the original (a resumed task) is left alone.
func releaseOriginal(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	sidecars *recreateSidecarsJob,
	name, originalId string,
) error {
	current, isFound, err := runtime.Inspect(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting "+name)
	}

	if !isFound || current.ID != originalId {
		return nil
	}

	backupName := name + sidecarBackupSuffix

	_, hasBackup, err := runtime.Inspect(ctx, backupName)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting "+backupName)
	}

	if hasBackup {
		return rerrors.Wrap(user_errors.ErrSidecarLeftoverExists, backupName)
	}

	err = sidecars.moveAside(ctx, runtime, name, backupName)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (j *switchRootJob) createNewRoot(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	plan switchRootPlan,
	oldRoot, newRoot container.InspectResponse,
) (string, error) {
	current, isFound, err := runtime.Inspect(ctx, plan.NewRootName)
	if err != nil {
		return "", rerrors.Wrap(err, "error inspecting "+plan.NewRootName)
	}

	if isFound && current.ID != newRoot.ID {
		err = runtime.Restart(ctx, current.ID)
		if err != nil {
			return "", rerrors.Wrap(err, "error starting the new network root")
		}

		return current.ID, nil
	}

	stamped := switchedRootLabels(j.payload.GetServiceName(), plan, oldRoot.ID)
	createReq := newRootCreateRequest(plan.NewRootName, stamped, oldRoot, newRoot)

	created, err := runtime.ContainerCreate(ctx, createReq)
	if err != nil {
		return "", rerrors.Wrap(err, "error creating the new network root")
	}

	err = runtime.Restart(ctx, created.ID)
	if err != nil {
		return "", rerrors.Wrap(err, "error starting the new network root")
	}

	err = attachMovedNetworks(ctx, runtime, created.ID, oldRoot)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	return created.ID, nil
}

func dropSwitchBackups(
	ctx context.Context, runtime container_runtime.ContainerRuntime, plan switchRootPlan,
) error {
	names := []string{plan.NewRootName}

	for _, name := range plan.Members {
		if name != plan.OldRootName {
			names = append(names, name)
		}
	}

	for _, name := range names {
		err := runtime.Remove(ctx, name+sidecarBackupSuffix)
		if err != nil {
			return rerrors.Wrap(err, "error dropping backup of "+name)
		}
	}

	return nil
}

// switchedRootLabels are the labels the new root is stamped with. It records the old root it replaces, which is
// the container onboarding leaves behind.
func switchedRootLabels(serviceName string, plan switchRootPlan, oldRootId string) map[string]string {
	stamped := sidecarLabels(serviceName)

	if plan.IsNewRootService {
		stamped = serviceMemberLabels(serviceName)
	}

	stamped[labels.OnboardedFromLabel] = oldRootId

	return stamped
}

// newRootCreateRequest clones source into the owner of the old root's network namespace: the old root's network
// mode, published ports, DNS settings, endpoint and aliases move to it.
func newRootCreateRequest(
	name string,
	stamped map[string]string,
	oldRoot, source container.InspectResponse,
) container_runtime.ContainerCreateRequest {
	createReq := memberCreateRequest(name, stamped, oldRoot.HostConfig.NetworkMode, source)

	createReq.Config.ExposedPorts = maps.Clone(oldRoot.Config.ExposedPorts)

	hostConfig := createReq.HostConfig.HostConfig

	hostConfig.PortBindings = movedPortBindings(oldRoot)
	hostConfig.PublishAllPorts = oldRoot.HostConfig.PublishAllPorts
	hostConfig.DNS = oldRoot.HostConfig.DNS
	hostConfig.DNSOptions = oldRoot.HostConfig.DNSOptions
	hostConfig.DNSSearch = oldRoot.HostConfig.DNSSearch
	hostConfig.ExtraHosts = oldRoot.HostConfig.ExtraHosts

	createReq.NetworkingConfig = movedNetworkingConfig(oldRoot)

	return createReq
}

// movedPortBindings are the host ports the old root actually holds, so a random mapping keeps its port.
func movedPortBindings(oldRoot container.InspectResponse) nat.PortMap {
	bindings := nat.PortMap{}

	if oldRoot.NetworkSettings != nil {
		for port, hostBindings := range oldRoot.NetworkSettings.Ports {
			if len(hostBindings) != 0 {
				bindings[port] = hostBindings
			}
		}
	}

	if len(bindings) == 0 {
		return maps.Clone(oldRoot.HostConfig.PortBindings)
	}

	return bindings
}

func primaryNetworkName(mode container.NetworkMode) string {
	if mode == "" || mode == defaultNetworkMode {
		return defaultBridgeNetwork
	}

	return string(mode)
}

func isEndpointNetwork(mode container.NetworkMode) bool {
	name := primaryNetworkName(mode)

	return name != hostNetwork && name != noneNetwork && !mode.IsContainer()
}

func isUserDefinedNetwork(name string) bool {
	return name != defaultBridgeNetwork && name != hostNetwork && name != noneNetwork
}

func movedNetworkingConfig(oldRoot container.InspectResponse) *container_runtime.NetworkingConfig {
	mode := oldRoot.HostConfig.NetworkMode

	if !isEndpointNetwork(mode) || oldRoot.NetworkSettings == nil {
		return nil
	}

	primary := primaryNetworkName(mode)

	endpoint, isFound := oldRoot.NetworkSettings.Networks[primary]
	if !isFound {
		return nil
	}

	settings := &network.EndpointSettings{}

	if isUserDefinedNetwork(primary) {
		settings.Aliases = movedAliases(oldRoot, endpoint)
	}

	config := &network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{primary: settings}}

	return &container_runtime.NetworkingConfig{NetworkingConfig: config}
}

// movedAliases are the aliases of the old root on a network without the ones Docker adds on its own, plus its
// container name, so peers still reach it by that name.
func movedAliases(oldRoot container.InspectResponse, endpoint *network.EndpointSettings) []string {
	aliases := make([]string, 0, len(endpoint.Aliases)+1)

	for _, alias := range endpoint.Aliases {
		isOwnShortId := strings.HasPrefix(oldRoot.ID, alias)
		if !isOwnShortId && !slices.Contains(aliases, alias) {
			aliases = append(aliases, alias)
		}
	}

	oldName := strings.TrimPrefix(oldRoot.Name, "/")
	if !slices.Contains(aliases, oldName) {
		aliases = append(aliases, oldName)
	}

	return aliases
}

// attachMovedNetworks connects the new root to the networks of the old root besides the one it was created in.
func attachMovedNetworks(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	newRootId string,
	oldRoot container.InspectResponse,
) error {
	if !isEndpointNetwork(oldRoot.HostConfig.NetworkMode) || oldRoot.NetworkSettings == nil {
		return nil
	}

	primary := primaryNetworkName(oldRoot.HostConfig.NetworkMode)

	for _, networkName := range slices.Sorted(maps.Keys(oldRoot.NetworkSettings.Networks)) {
		if networkName == primary {
			continue
		}

		endpoint := oldRoot.NetworkSettings.Networks[networkName]

		attachReq := container_runtime.AttachContainerRequest{
			ContainerID: newRootId,
			NetworkId:   endpoint.NetworkID,
		}

		if isUserDefinedNetwork(networkName) {
			attachReq.Aliases = movedAliases(oldRoot, endpoint)
		}

		err := runtime.AttachContainer(ctx, attachReq)
		if err != nil {
			return rerrors.Wrap(err, "error connecting the new network root to "+networkName)
		}
	}

	return nil
}
