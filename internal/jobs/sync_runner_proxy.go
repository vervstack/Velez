package jobs

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers"
)

const (
	stepSyncRunnerProxy = "sync_runner_proxy"
)

// syncRunnerProxyJob mirrors the proxy env of the container into its runner's job environment -
// see providers.Provider.SyncProxy - and into the buildkitd inside the runner's DinD, when it has one.
// It is a no-op for a container that is not a runner.
type syncRunnerProxyJob struct {
	runtimes container_runtime.RuntimeResolver

	req           upgradeRequestAccessor
	containerName string
}

func (j *syncRunnerProxyJob) Do(ctx context.Context) error {
	err := j.sync(ctx, j.containerName)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (j *syncRunnerProxyJob) sync(ctx context.Context, name string) error {
	runtime, err := j.runtimes.Runtime(ctx, j.req.GetUpgradeRequest().GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	info, isFound, err := runtime.Inspect(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting upgraded container")
	}

	if !isFound || info.Config == nil || info.Config.Labels[labels.RunnerInstanceLabel] != labelTrueValue {
		return nil
	}

	providerName := info.Config.Labels[labels.RunnerProviderLabel]
	providerEnum := velez_api.RunnerProvider(velez_api.RunnerProvider_value[providerName])

	runnerProvider, err := providers.For(providerEnum)
	if err != nil {
		return rerrors.Wrap(err, "error resolving runner provider")
	}

	err = runnerProvider.SyncProxy(ctx, runtime, name)
	if err != nil {
		return rerrors.Wrap(err, "error syncing runner proxy")
	}

	dindName := info.Config.Labels[labels.RunnerDindLabel]
	if dindName == "" || providerEnum != velez_api.RunnerProvider_GITLAB {
		return nil
	}

	facts := runnerBuildkitFacts{
		name:        name,
		environment: j.req.GetUpgradeRequest().GetEnvironment(),
		provider:    providerEnum,
		dindName:    dindName,
	}

	err = syncBuildkitProxy(ctx, j.runtimes, runtime, facts, info.Config.Env)
	if err != nil {
		return rerrors.Wrap(err, "error syncing buildkit proxy")
	}

	return nil
}
