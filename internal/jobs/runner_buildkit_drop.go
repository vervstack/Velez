package jobs

import (
	"context"
	"slices"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/proxyenv"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers"
	"go.vervstack.ru/Velez/internal/storage"
)

// setBuildkitNetworkModeJob points the runner's job containers at the BuildKit
// network, or - when not enabled - away from it.
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

	var networkMode string

	if j.isEnabled {
		buildkitNetwork, isFound, findErr := findBuildkitNetwork(ctx, runtime, facts.name)
		if findErr != nil {
			return findErr
		}

		if !isFound {
			return rerrors.Wrap(errBuildkitNetworkMissing, facts.name)
		}

		networkMode = buildkitNetwork.DockerName
	}

	err = runnerProvider.ApplyNetworkMode(ctx, runtime, facts.name, networkMode)
	if err != nil {
		return rerrors.Wrap(err, "error applying runner network mode")
	}

	return nil
}

type dropBuildkitSidecarJob struct {
	dataStorage storage.Storage
	runtimes    container_runtime.RuntimeResolver

	target runnerBuildkitTarget
}

func (j *dropBuildkitSidecarJob) Do(ctx context.Context) error {
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
		err = runtime.Remove(ctx, sidecarName)
		if err != nil {
			return rerrors.Wrap(err, "error removing buildkit sidecar")
		}
	}

	if !j.dataStorage.IsStatefull() {
		return nil
	}

	bindings := j.dataStorage.ContainerBindings()
	if bindings == nil {
		return nil
	}

	err = bindings.Delete(ctx, domain.SelfNodeId, facts.environment, sidecarName)
	if err != nil {
		return rerrors.Wrap(err, "error deleting buildkit sidecar binding")
	}

	return nil
}

type removeBuildkitNetworkJob struct {
	runtimes container_runtime.RuntimeResolver

	target runnerBuildkitTarget
}

func (j *removeBuildkitNetworkJob) Do(ctx context.Context) error {
	facts, runtime, err := resolveBuildkitRuntime(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	buildkitNetwork, isFound, err := findBuildkitNetwork(ctx, runtime, facts.name)
	if err != nil {
		return err
	}

	if !isFound {
		return nil
	}

	err = runtime.RemoveNetwork(ctx, buildkitNetwork.Id)
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
	facts, runtime, err := resolveBuildkitRuntime(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	err = runtime.RemoveVolume(ctx, domain.RunnerBuildkitStateVolumeName(facts.name))
	if err != nil {
		return rerrors.Wrap(err, "error removing buildkit state volume")
	}

	return nil
}

// runnerBuildkitNetworkMode is the network the runner's job containers have to
// join: the BuildKit network when the runner has a sidecar, empty when it has
// none. Used where the runner's config.toml is rewritten and the setting has to
// survive it.
func runnerBuildkitNetworkMode(
	ctx context.Context, runtime container_runtime.ContainerRuntime, runnerName string,
) (string, error) {
	sidecar, isFound, err := runtime.Inspect(ctx, domain.RunnerBuildkitServiceName(runnerName))
	if err != nil {
		return "", rerrors.Wrap(err, "error inspecting buildkit sidecar")
	}

	if !isFound || sidecar.Config == nil || sidecar.Config.Labels[labels.BuildkitForLabel] != runnerName {
		return "", nil
	}

	buildkitNetwork, isNetworkFound, err := findBuildkitNetwork(ctx, runtime, runnerName)
	if err != nil {
		return "", err
	}

	if !isNetworkFound {
		return "", nil
	}

	return buildkitNetwork.DockerName, nil
}

// syncBuildkitProxy gives the runner's BuildKit sidecar the proxy env of the
// runner container, recreating the sidecar when it differs - Docker cannot
// change the env of a created container. A runner without a sidecar is left
// alone.
func syncBuildkitProxy(
	ctx context.Context, runtime container_runtime.ContainerRuntime, runnerName string, runnerEnv []string,
) error {
	sidecarName := domain.RunnerBuildkitServiceName(runnerName)

	sidecar, isFound, err := runtime.Inspect(ctx, sidecarName)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting buildkit sidecar")
	}

	if !isFound || sidecar.Config == nil || sidecar.Config.Labels[labels.BuildkitForLabel] != runnerName {
		return nil
	}

	wantProxy := buildkitProxyEnv(runnerEnv)

	if slices.Equal(wantProxy, proxyEntries(sidecar.Config.Env)) {
		return nil
	}

	createReq := recreatedBuildkitSidecarRequest(sidecarName, sidecar, wantProxy)

	err = runtime.Remove(ctx, sidecarName)
	if err != nil {
		return rerrors.Wrap(err, "error removing buildkit sidecar")
	}

	created, err := runtime.ContainerCreate(ctx, createReq)
	if err != nil {
		return rerrors.Wrap(err, "error recreating buildkit sidecar")
	}

	err = runtime.Restart(ctx, created.ID)
	if err != nil {
		return rerrors.Wrap(err, "error starting buildkit sidecar")
	}

	return nil
}

// recreatedBuildkitSidecarRequest clones an inspected sidecar into a create
// request, with its proxy env replaced by wantProxy.
func recreatedBuildkitSidecarRequest(
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
