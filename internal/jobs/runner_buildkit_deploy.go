package jobs

import (
	"context"
	"fmt"
	"slices"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/proxyenv"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	buildkitDescriptorName = "buildkit"
	binfmtDescriptorName   = "binfmt"

	buildkitStatePath = "/var/lib/buildkit"
	buildkitAddrFlag  = "--addr"

	binfmtNameSuffix  = "-binfmt"
	binfmtInstallFlag = "--install"
	binfmtAllTargets  = "all"
)

// buildkitListenAddress is what buildkitd listens on inside the private network.
func buildkitListenAddress() string {
	return fmt.Sprintf("tcp://0.0.0.0:%d", domain.RunnerBuildkitPort)
}

func buildkitBinfmtName(runnerName string) string {
	return domain.RunnerBuildkitServiceName(runnerName) + binfmtNameSuffix
}

// resolveBuildkitRuntime locates the target runner and the container runtime of
// its environment.
func resolveBuildkitRuntime(
	ctx context.Context, runtimes container_runtime.RuntimeResolver, target runnerBuildkitTarget,
) (runnerBuildkitFacts, container_runtime.ContainerRuntime, error) {
	facts, err := target.Locate(ctx)
	if err != nil {
		return runnerBuildkitFacts{}, nil, rerrors.Wrap(err, "error locating runner")
	}

	runtime, err := runtimes.Runtime(ctx, facts.environment)
	if err != nil {
		return runnerBuildkitFacts{}, nil, rerrors.Wrap(err, "error resolving container runtime")
	}

	return facts, runtime, nil
}

// findBuildkitNetwork looks up the runner's BuildKit network by its logical name.
func findBuildkitNetwork(
	ctx context.Context, runtime container_runtime.ContainerRuntime, runnerName string,
) (container_runtime.NetworkInfo, bool, error) {
	networkName := domain.RunnerBuildkitServiceName(runnerName)

	networks, err := runtime.ListNetworks(ctx, true)
	if err != nil {
		return container_runtime.NetworkInfo{}, false, rerrors.Wrap(err, "error listing networks")
	}

	for _, found := range networks {
		if found.Name == networkName {
			return found, true, nil
		}
	}

	return container_runtime.NetworkInfo{}, false, nil
}

type ensureBuildkitNetworkJob struct {
	runtimes container_runtime.RuntimeResolver

	target runnerBuildkitTarget
}

func (j *ensureBuildkitNetworkJob) Do(ctx context.Context) error {
	facts, runtime, err := resolveBuildkitRuntime(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	networkName := domain.RunnerBuildkitServiceName(facts.name)

	err = runtime.CreateNetwork(ctx, networkName)
	if err != nil {
		return rerrors.Wrapf(err, "error creating network: %s", networkName)
	}

	return nil
}

// installBinfmtJob registers the qemu handlers multi-arch builds need. The
// handlers are kernel-global, so running the installer again changes nothing.
type installBinfmtJob struct {
	boxes    vervonomicon.BoxLookup
	runtimes container_runtime.RuntimeResolver

	target runnerBuildkitTarget
}

func (j *installBinfmtJob) Do(ctx context.Context) error {
	facts, runtime, err := resolveBuildkitRuntime(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	_, smerdRequest, err := resolveS3Descriptor(
		ctx, j.boxes, binfmtDescriptorName, facts.environment, "", nil, nil, nil)
	if err != nil {
		return rerrors.Wrap(err, "error building binfmt installer request")
	}

	_, err = runtime.PullImage(ctx, smerdRequest.GetImageName())
	if err != nil {
		return rerrors.Wrap(err, "error pulling binfmt installer image")
	}

	installerName := buildkitBinfmtName(facts.name)

	err = runtime.Remove(ctx, installerName)
	if err != nil {
		return rerrors.Wrap(err, "error removing leftover binfmt installer")
	}

	createReq := newBinfmtInstallerRequest(installerName, smerdRequest.GetImageName())

	created, err := runtime.ContainerCreate(ctx, createReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating binfmt installer")
	}

	err = runtime.Restart(ctx, created.ID)
	if err != nil {
		return rerrors.Wrap(err, "error starting binfmt installer")
	}

	var exitCode int

	probe := binfmtProbe{runtime: runtime, name: installerName, exitCode: &exitCode}

	err = pollGarage(ctx, probe.check, "timed out waiting for binfmt installer to finish")
	if err != nil {
		return err
	}

	err = runtime.Remove(ctx, installerName)
	if err != nil {
		return rerrors.Wrap(err, "error removing binfmt installer")
	}

	if exitCode != 0 {
		return rerrors.Wrap(errBinfmtInstallFailed)
	}

	return nil
}

func newBinfmtInstallerRequest(name, image string) container_runtime.ContainerCreateRequest {
	config := &container.Config{
		Image:  image,
		Cmd:    strslice.StrSlice{binfmtInstallFlag, binfmtAllTargets},
		Labels: map[string]string{labels.Sidecar: labelTrueValue},
	}

	hostConfig := &container.HostConfig{Privileged: true}

	return container_runtime.ContainerCreateRequest{
		Config:        &container_runtime.ContainerConfig{Config: config},
		HostConfig:    &container_runtime.HostConfig{HostConfig: hostConfig},
		ContainerName: name,
	}
}

type binfmtProbe struct {
	runtime  container_runtime.ContainerRuntime
	name     string
	exitCode *int
}

func (p binfmtProbe) check(ctx context.Context) error {
	info, isFound, err := p.runtime.Inspect(ctx, p.name)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting binfmt installer")
	}

	if !isFound {
		return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound, p.name)
	}

	if info.State == nil || info.State.Status != container.StateExited {
		return rerrors.Wrap(errBinfmtNotFinished)
	}

	*p.exitCode = info.State.ExitCode

	return nil
}

// buildkitSidecarSpec is everything the BuildKit sidecar container is built from.
type buildkitSidecarSpec struct {
	runnerName        string
	imageName         string
	networkDockerName string
	restart           container.RestartPolicy
	env               []string
	isPrivileged      bool
}

func newBuildkitSidecarRequest(spec buildkitSidecarSpec) container_runtime.ContainerCreateRequest {
	containerLabels := sidecarLabels(spec.runnerName)

	containerLabels[labels.ComposeGroupLabel] = spec.runnerName
	containerLabels[labels.BuildkitForLabel] = spec.runnerName

	config := &container.Config{
		Image:  spec.imageName,
		Cmd:    strslice.StrSlice{buildkitAddrFlag, buildkitListenAddress()},
		Env:    spec.env,
		Labels: containerLabels,
	}

	stateBind := domain.RunnerBuildkitStateVolumeName(spec.runnerName) + ":" + buildkitStatePath

	hostConfig := &container.HostConfig{
		NetworkMode:   container.NetworkMode(spec.networkDockerName),
		RestartPolicy: spec.restart,
		Privileged:    spec.isPrivileged,
		Binds:         []string{stateBind},
	}

	endpoint := &network.EndpointSettings{Aliases: []string{domain.RunnerBuildkitAlias}}

	networking := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{spec.networkDockerName: endpoint},
	}

	return container_runtime.ContainerCreateRequest{
		Config:           &container_runtime.ContainerConfig{Config: config},
		HostConfig:       &container_runtime.HostConfig{HostConfig: hostConfig},
		NetworkingConfig: &container_runtime.NetworkingConfig{NetworkingConfig: networking},
		ContainerName:    domain.RunnerBuildkitServiceName(spec.runnerName),
	}
}

// buildkitProxyEnv is the proxy env of the runner container, as `KEY=value`
// entries; empty when the runner has no proxy.
func buildkitProxyEnv(runnerEnv []string) []string {
	proxyUrl, bypassHosts := proxyenv.ParseList(runnerEnv)
	entries := parser.FromDockerEnv(proxyenv.Env(proxyUrl, bypassHosts))

	slices.Sort(entries)

	return entries
}

type deployBuildkitJob struct {
	boxes    vervonomicon.BoxLookup
	settings storage.SettingsStorage
	runtimes container_runtime.RuntimeResolver

	target runnerBuildkitTarget
}

func (j *deployBuildkitJob) Do(ctx context.Context) error {
	facts, runtime, err := resolveBuildkitRuntime(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	sidecarName := domain.RunnerBuildkitServiceName(facts.name)

	existing, isFound, err := runtime.Inspect(ctx, sidecarName)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting buildkit sidecar")
	}

	if isFound && existing.Config != nil && existing.Config.Labels[labels.BuildkitForLabel] == facts.name {
		return j.startExisting(ctx, runtime, existing)
	}

	_, smerdRequest, err := resolveS3Descriptor(
		ctx, j.boxes, buildkitDescriptorName, facts.environment, "", nil, nil, nil)
	if err != nil {
		return rerrors.Wrap(err, "error building buildkit deploy request")
	}

	_, err = runtime.PullImage(ctx, smerdRequest.GetImageName())
	if err != nil {
		return rerrors.Wrap(err, "error pulling buildkit image")
	}

	spec, err := j.buildSpec(ctx, runtime, facts, smerdRequest.GetImageName())
	if err != nil {
		return err
	}

	createReq := newBuildkitSidecarRequest(spec)

	created, err := runtime.ContainerCreate(ctx, createReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating buildkit sidecar")
	}

	err = runtime.Restart(ctx, created.ID)
	if err != nil {
		return rerrors.Wrap(err, "error starting buildkit sidecar")
	}

	return nil
}

// startExisting starts a sidecar an earlier, interrupted run created, leaving a
// running one alone.
func (j *deployBuildkitJob) startExisting(
	ctx context.Context, runtime container_runtime.ContainerRuntime, existing container.InspectResponse,
) error {
	if existing.State != nil && existing.State.Running {
		return nil
	}

	err := runtime.Restart(ctx, existing.ID)
	if err != nil {
		return rerrors.Wrap(err, "error starting buildkit sidecar")
	}

	return nil
}

// buildSpec reads what the sidecar inherits from the runner container: its
// restart policy and its proxy; the isolation follows the node's Sysbox setting,
// like a DinD service.
func (j *deployBuildkitJob) buildSpec(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	facts runnerBuildkitFacts,
	imageName string,
) (buildkitSidecarSpec, error) {
	runner, isFound, err := runtime.Inspect(ctx, facts.name)
	if err != nil {
		return buildkitSidecarSpec{}, rerrors.Wrap(err, "error inspecting runner container")
	}

	if !isFound {
		return buildkitSidecarSpec{}, rerrors.Wrap(user_errors.ErrRegisterContainerNotFound, facts.name)
	}

	buildkitNetwork, isNetworkFound, err := findBuildkitNetwork(ctx, runtime, facts.name)
	if err != nil {
		return buildkitSidecarSpec{}, err
	}

	if !isNetworkFound {
		return buildkitSidecarSpec{}, rerrors.Wrap(errBuildkitNetworkMissing, facts.name)
	}

	settings, err := j.settings.GetSettings(ctx)
	if err != nil {
		return buildkitSidecarSpec{}, rerrors.Wrap(err, "error getting node settings")
	}

	spec := buildkitSidecarSpec{
		runnerName:        facts.name,
		imageName:         imageName,
		networkDockerName: buildkitNetwork.DockerName,
		isPrivileged:      !settings.IsSysboxEnabled,
	}

	if runner.HostConfig != nil {
		spec.restart = runner.HostConfig.RestartPolicy
	}

	if runner.Config != nil {
		spec.env = buildkitProxyEnv(runner.Config.Env)
	}

	return spec, nil
}

type waitBuildkitJob struct {
	runtimes container_runtime.RuntimeResolver

	target runnerBuildkitTarget
}

func (j *waitBuildkitJob) Do(ctx context.Context) error {
	facts, runtime, err := resolveBuildkitRuntime(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	probe := buildkitRunningProbe{runtime: runtime, name: domain.RunnerBuildkitServiceName(facts.name)}

	return pollGarage(ctx, probe.check, "timed out waiting for buildkit to run")
}

type buildkitRunningProbe struct {
	runtime container_runtime.ContainerRuntime
	name    string
}

func (p buildkitRunningProbe) check(ctx context.Context) error {
	isRunning, isFound, err := p.runtime.IsContainerRunning(ctx, p.name)
	if err != nil {
		return rerrors.Wrap(err, "error checking container state")
	}

	if !isFound || !isRunning {
		return rerrors.Wrap(errBuildkitNotRunning)
	}

	return nil
}

// registerBuildkitBindingJob binds the sidecar to the runner's service in
// cluster mode, where the binding is what marks the runner as having BuildKit.
// Single-node mode derives that from the sidecar's labels, so there is nothing
// to write.
type registerBuildkitBindingJob struct {
	dataStorage storage.Storage

	target runnerBuildkitTarget
}

func (j *registerBuildkitBindingJob) Do(ctx context.Context) error {
	if !j.dataStorage.IsStatefull() {
		return nil
	}

	facts, err := j.target.Locate(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error locating runner")
	}

	bindings := j.dataStorage.ContainerBindings()
	if bindings == nil {
		return rerrors.Wrap(user_errors.ErrContainerBindingsUnavailable)
	}

	svc, err := j.dataStorage.Services().GetByName(ctx, facts.name)
	if err != nil {
		return rerrors.Wrap(err, "error getting runner service")
	}

	binding := domain.ContainerBinding{
		ServiceId:     svc.ID,
		ServiceName:   facts.name,
		NodeId:        domain.SelfNodeId,
		Environment:   facts.environment,
		ContainerName: domain.RunnerBuildkitServiceName(facts.name),
		IsSidecar:     true,
	}

	err = bindings.Upsert(ctx, binding)
	if err != nil {
		return rerrors.Wrap(err, "error upserting buildkit sidecar binding")
	}

	return nil
}
