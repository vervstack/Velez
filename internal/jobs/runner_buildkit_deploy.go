package jobs

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/proxyenv"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
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

// openBuildkitDind locates the runner and opens the runtime of the DinD daemon
// buildkitd runs in. The io.Closer releases the connection to that daemon.
func openBuildkitDind(
	ctx context.Context, runtimes container_runtime.RuntimeResolver, target runnerBuildkitTarget,
) (runnerBuildkitFacts, container_runtime.ContainerRuntime, io.Closer, error) {
	facts, err := target.Locate(ctx)
	if err != nil {
		return runnerBuildkitFacts{}, nil, nil, rerrors.Wrap(err, "error locating runner")
	}

	if facts.dindName == "" || facts.provider != velez_api.RunnerProvider_GITLAB {
		return runnerBuildkitFacts{}, nil, nil, rerrors.Wrap(user_errors.ErrRunnerBuildkitRequiresDind, facts.name)
	}

	dind, closer, err := runtimes.NestedRuntime(ctx, facts.environment, facts.dindName)
	if err != nil {
		return runnerBuildkitFacts{}, nil, nil, rerrors.Wrapf(err, "error opening dind daemon: %s", facts.dindName)
	}

	return facts, dind, closer, nil
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
	facts, dind, closer, err := openBuildkitDind(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	defer common.CloseWithLog(closer.Close, "buildkit dind client")

	networkName := domain.RunnerBuildkitServiceName(facts.name)

	err = dind.CreateNetwork(ctx, networkName)
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

// buildkitdSpec is everything the buildkitd container inside the DinD is built from.
type buildkitdSpec struct {
	runnerName string
	imageName  string
	env        []string
}

func newBuildkitdRequest(spec buildkitdSpec) container_runtime.ContainerCreateRequest {
	config := &container.Config{
		Image: spec.imageName,
		Cmd:   strslice.StrSlice{buildkitAddrFlag, buildkitListenAddress()},
		Env:   spec.env,
	}

	networkName := domain.RunnerBuildkitServiceName(spec.runnerName)
	stateBind := domain.RunnerBuildkitStateVolumeName(spec.runnerName) + ":" + buildkitStatePath

	// Privileged inside the DinD: buildkitd needs it for its rootful overlay and
	// qemu-user builds, and the DinD is the isolation boundary.
	hostConfig := &container.HostConfig{
		NetworkMode:   container.NetworkMode(networkName),
		RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyUnlessStopped},
		Privileged:    true,
		Binds:         []string{stateBind},
	}

	endpoint := &network.EndpointSettings{Aliases: []string{domain.RunnerBuildkitAlias}}

	networking := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{networkName: endpoint},
	}

	return container_runtime.ContainerCreateRequest{
		Config:           &container_runtime.ContainerConfig{Config: config},
		HostConfig:       &container_runtime.HostConfig{HostConfig: hostConfig},
		NetworkingConfig: &container_runtime.NetworkingConfig{NetworkingConfig: networking},
		ContainerName:    networkName,
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
	runtimes container_runtime.RuntimeResolver

	target runnerBuildkitTarget
}

func (j *deployBuildkitJob) Do(ctx context.Context) error {
	facts, dind, closer, err := openBuildkitDind(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	defer common.CloseWithLog(closer.Close, "buildkit dind client")

	buildkitdName := domain.RunnerBuildkitServiceName(facts.name)

	existing, isFound, err := dind.Inspect(ctx, buildkitdName)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting buildkitd")
	}

	if isFound {
		return startExistingBuildkitd(ctx, dind, existing)
	}

	_, smerdRequest, err := resolveS3Descriptor(
		ctx, j.boxes, buildkitDescriptorName, facts.environment, "", nil, nil, nil)
	if err != nil {
		return rerrors.Wrap(err, "error building buildkit deploy request")
	}

	_, err = dind.PullImage(ctx, smerdRequest.GetImageName())
	if err != nil {
		return rerrors.Wrap(err, "error pulling buildkit image")
	}

	spec, err := j.buildSpec(ctx, facts, smerdRequest.GetImageName())
	if err != nil {
		return err
	}

	createReq := newBuildkitdRequest(spec)

	created, err := dind.ContainerCreate(ctx, createReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating buildkitd")
	}

	err = dind.Restart(ctx, created.ID)
	if err != nil {
		return rerrors.Wrap(err, "error starting buildkitd")
	}

	return nil
}

// startExistingBuildkitd starts a buildkitd an earlier, interrupted run
// created, leaving a running one alone.
func startExistingBuildkitd(
	ctx context.Context, dind container_runtime.ContainerRuntime, existing container.InspectResponse,
) error {
	if existing.State != nil && existing.State.Running {
		return nil
	}

	err := dind.Restart(ctx, existing.ID)
	if err != nil {
		return rerrors.Wrap(err, "error starting buildkitd")
	}

	return nil
}

// buildSpec reads what buildkitd inherits from the runner container: its proxy.
func (j *deployBuildkitJob) buildSpec(
	ctx context.Context, facts runnerBuildkitFacts, imageName string,
) (buildkitdSpec, error) {
	node, err := j.runtimes.Runtime(ctx, facts.environment)
	if err != nil {
		return buildkitdSpec{}, rerrors.Wrap(err, "error resolving container runtime")
	}

	runner, isFound, err := node.Inspect(ctx, facts.name)
	if err != nil {
		return buildkitdSpec{}, rerrors.Wrap(err, "error inspecting runner container")
	}

	if !isFound {
		return buildkitdSpec{}, rerrors.Wrap(user_errors.ErrRegisterContainerNotFound, facts.name)
	}

	spec := buildkitdSpec{runnerName: facts.name, imageName: imageName}

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
	facts, dind, closer, err := openBuildkitDind(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	defer common.CloseWithLog(closer.Close, "buildkit dind client")

	probe := buildkitRunningProbe{runtime: dind, name: domain.RunnerBuildkitServiceName(facts.name)}

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

// markBuildkitJob records whether the runner has BuildKit, in cluster mode where
// the runners row is the source of truth. Single-node mode derives the flag from
// the runner's config.toml, so the storage write is a no-op there.
type markBuildkitJob struct {
	services storage.ServicesStorage
	runners  storage.RunnersStorage

	target    runnerBuildkitTarget
	isEnabled bool
}

func (j *markBuildkitJob) Do(ctx context.Context) error {
	facts, err := j.target.Locate(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error locating runner")
	}

	svc, err := j.services.GetByName(ctx, facts.name)
	if err != nil {
		return rerrors.Wrap(err, "error getting runner service")
	}

	err = j.runners.SetRunnerBuildkit(ctx, svc.ID, j.isEnabled)
	if err != nil {
		return rerrors.Wrap(err, "error recording runner buildkit state")
	}

	return nil
}
