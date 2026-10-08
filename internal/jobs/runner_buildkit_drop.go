package jobs

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/gitlab_runner_config"
	"go.vervstack.ru/Velez/internal/proxyenv"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers"
	"go.vervstack.ru/Velez/internal/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
)

// setBuildkitNetworkModeJob points the runner's job containers at the BuildKit
// network of the DinD they run in, or - when not enabled - away from it.
type setBuildkitNetworkModeJob struct {
	runtimes container_runtime.RuntimeResolver

	target    runnerBuildkitTarget
	isEnabled bool
}

func (j *setBuildkitNetworkModeJob) Do(ctx context.Context) error {
	facts, runtime, err := resolveBuildkitRuntime(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	runnerProvider, err := providers.For(facts.provider)
	if err != nil {
		return rerrors.Wrap(err, "error resolving runner provider")
	}

	networkMode := buildkitNetworkMode(j.isEnabled, facts.name)

	err = runnerProvider.ApplyNetworkMode(ctx, runtime, facts.name, networkMode)
	if err != nil {
		return rerrors.Wrap(err, "error applying runner network mode")
	}

	return nil
}

// buildkitNetworkMode is the docker network the runner's job containers join:
// the BuildKit network of the DinD when enabled, none otherwise.
func buildkitNetworkMode(isEnabled bool, runnerName string) string {
	if !isEnabled {
		return ""
	}

	return domain.RunnerBuildkitServiceName(runnerName)
}

type removeBuildkitJob struct {
	runtimes container_runtime.RuntimeResolver

	target runnerBuildkitTarget
}

func (j *removeBuildkitJob) Do(ctx context.Context) error {
	facts, dind, closer, err := openBuildkitDind(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	defer common.CloseWithLog(closer.Close, "buildkit dind client")

	return removeBuildkitd(ctx, dind, facts.name)
}

func removeBuildkitd(ctx context.Context, dind container_runtime.ContainerRuntime, runnerName string) error {
	err := dind.Remove(ctx, domain.RunnerBuildkitServiceName(runnerName))
	if err != nil {
		return rerrors.Wrap(err, "error removing buildkitd")
	}

	return nil
}

type removeBuildkitNetworkJob struct {
	runtimes container_runtime.RuntimeResolver

	target runnerBuildkitTarget
}

func (j *removeBuildkitNetworkJob) Do(ctx context.Context) error {
	facts, dind, closer, err := openBuildkitDind(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	defer common.CloseWithLog(closer.Close, "buildkit dind client")

	return removeBuildkitNetwork(ctx, dind, facts.name)
}

func removeBuildkitNetwork(ctx context.Context, dind container_runtime.ContainerRuntime, runnerName string) error {
	buildkitNetwork, isFound, err := findBuildkitNetwork(ctx, dind, runnerName)
	if err != nil {
		return err
	}

	if !isFound {
		return nil
	}

	err = dind.RemoveNetwork(ctx, buildkitNetwork.Id)
	if err != nil {
		return rerrors.Wrapf(err, "error removing network: %s", buildkitNetwork.Name)
	}

	return nil
}

type removeBuildkitVolumeJob struct {
	runtimes container_runtime.RuntimeResolver

	target runnerBuildkitTarget
}

func (j *removeBuildkitVolumeJob) Do(ctx context.Context) error {
	facts, dind, closer, err := openBuildkitDind(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	defer common.CloseWithLog(closer.Close, "buildkit dind client")

	return removeBuildkitVolume(ctx, dind, facts.name)
}

func removeBuildkitVolume(ctx context.Context, dind container_runtime.ContainerRuntime, runnerName string) error {
	err := dind.RemoveVolume(ctx, domain.RunnerBuildkitStateVolumeName(runnerName))
	if err != nil {
		return rerrors.Wrap(err, "error removing buildkit state volume")
	}

	return nil
}

// isRunnerBuildkitNetworkModeSet reads the runner's config.toml: BuildKit counts
// as enabled for the runner once its job containers are pointed at the BuildKit
// network. A runner without a config.toml has none.
func isRunnerBuildkitNetworkModeSet(
	ctx context.Context, runtime container_runtime.ContainerRuntime, runnerName string,
) (bool, error) {
	config, err := runtime.CopyFromContainer(ctx, runnerName, gitlab_runner_config.ConfigPath)
	if errors.Is(err, user_errors.ErrContainerFileNotFound) || errdefs.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, rerrors.Wrap(err, "error reading gitlab-runner config.toml")
	}

	networkMode, err := gitlab_runner_config.NetworkMode(config)
	if err != nil {
		return false, rerrors.Wrap(err)
	}

	return networkMode == domain.RunnerBuildkitServiceName(runnerName), nil
}

// syncBuildkitProxy gives the runner's buildkitd inside its DinD the proxy env of
// the runner container, recreating buildkitd when it differs - Docker cannot
// change the env of a created container. A runner without BuildKit is left alone.
func syncBuildkitProxy(
	ctx context.Context,
	runtimes container_runtime.RuntimeResolver,
	node container_runtime.ContainerRuntime,
	facts runnerBuildkitFacts,
	runnerEnv []string,
) error {
	isEnabled, err := isRunnerBuildkitNetworkModeSet(ctx, node, facts.name)
	if err != nil {
		return err
	}

	if !isEnabled {
		return nil
	}

	dind, closer, err := runtimes.NestedRuntime(ctx, facts.environment, facts.dindName)
	if err != nil {
		return rerrors.Wrapf(err, "error opening dind daemon: %s", facts.dindName)
	}

	defer common.CloseWithLog(closer.Close, "buildkit dind client")

	buildkitdName := domain.RunnerBuildkitServiceName(facts.name)

	buildkitd, isFound, err := dind.Inspect(ctx, buildkitdName)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting buildkitd")
	}

	if !isFound || buildkitd.Config == nil || buildkitd.HostConfig == nil {
		return nil
	}

	wantProxy := buildkitProxyEnv(runnerEnv)

	if slices.Equal(wantProxy, proxyEntries(buildkitd.Config.Env)) {
		return nil
	}

	createReq := recreatedBuildkitdRequest(buildkitdName, buildkitd, wantProxy)

	err = dind.Remove(ctx, buildkitdName)
	if err != nil {
		return rerrors.Wrap(err, "error removing buildkitd")
	}

	created, err := dind.ContainerCreate(ctx, createReq)
	if err != nil {
		return rerrors.Wrap(err, "error recreating buildkitd")
	}

	err = dind.Restart(ctx, created.ID)
	if err != nil {
		return rerrors.Wrap(err, "error starting buildkitd")
	}

	return nil
}

// recreatedBuildkitdRequest clones an inspected buildkitd into a create
// request, with its proxy env replaced by wantProxy.
func recreatedBuildkitdRequest(
	name string, source container.InspectResponse, wantProxy []string,
) container_runtime.ContainerCreateRequest {
	config := *source.Config

	config.Hostname = ""
	config.Domainname = ""

	config.Env = append(withoutProxyEntries(source.Config.Env), wantProxy...)

	hostConfig := *source.HostConfig

	networkName := string(hostConfig.NetworkMode)
	endpoint := &network.EndpointSettings{Aliases: []string{domain.RunnerBuildkitAlias}}

	networking := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{networkName: endpoint},
	}

	return container_runtime.ContainerCreateRequest{
		Config:           &container_runtime.ContainerConfig{Config: &config},
		HostConfig:       &container_runtime.HostConfig{HostConfig: &hostConfig},
		NetworkingConfig: &container_runtime.NetworkingConfig{NetworkingConfig: networking},
		ContainerName:    name,
	}
}

func isProxyEntry(entry string) bool {
	key, _, _ := strings.Cut(entry, "=")

	return slices.Contains(proxyenv.Keys(), key)
}

// proxyEntries are the proxy `KEY=value` entries of env, sorted.
func proxyEntries(env []string) []string {
	entries := make([]string, 0, len(proxyenv.Keys()))

	for _, entry := range env {
		if isProxyEntry(entry) {
			entries = append(entries, entry)
		}
	}

	slices.Sort(entries)

	return entries
}

func withoutProxyEntries(env []string) []string {
	kept := make([]string, 0, len(env))

	for _, entry := range env {
		if !isProxyEntry(entry) {
			kept = append(kept, entry)
		}
	}

	return kept
}
