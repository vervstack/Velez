package jobs

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	stepWaitForRunnerRedeploy = "wait_for_runner_redeploy"

	releaseBuildkitTimeout   = 30 * time.Second
	runnerRedeployWait       = 5 * time.Minute
	runnerRedeployPollPeriod = 2 * time.Second
)

// ReleaseRunnerBuildkitReq names a runner that is about to leave a DinD that
// runs its BuildKit.
type ReleaseRunnerBuildkitReq struct {
	RunnerName  string
	Environment string
	Provider    velez_api.RunnerProvider
	// DindName is the DinD the runner leaves - where buildkitd runs now.
	DindName string
	// IsDetach is set when the runner leaves for no DinD at all: BuildKit needs
	// a DinD, so it is turned off for good. Otherwise it stays enabled, to be
	// deployed on the new DinD by the runner's redeploy.
	IsDetach bool
}

// ReleaseRunnerBuildkit removes the runner's buildkitd, network and state volume
// from the DinD it leaves. Removing is best-effort - the DinD may be stopped or
// gone, and that must not block the move. Turning BuildKit off on detach is not:
// until the runner's marker and config.toml say off, its jobs would keep pointing at
// a network that no longer exists.
func ReleaseRunnerBuildkit(
	ctx context.Context,
	dataStorage storage.Storage,
	runtimes container_runtime.RuntimeResolver,
	req ReleaseRunnerBuildkitReq,
) error {
	facts := runnerBuildkitFacts{
		name:        req.RunnerName,
		environment: req.Environment,
		provider:    req.Provider,
		dindName:    req.DindName,
	}

	if req.IsDetach {
		steps := runnerBuildkitJobs{dataStorage: dataStorage, runtimes: runtimes}

		for _, step := range steps.turnOff(facts) {
			err := step.Job.Do(ctx)
			if err != nil {
				return rerrors.Wrapf(err, "error running %s", step.Name)
			}
		}
	}

	if req.DindName == "" {
		return nil
	}

	removeCtx, cancel := context.WithTimeout(ctx, releaseBuildkitTimeout)
	defer cancel()

	err := removeBuildkitFromDind(removeCtx, runtimes, req.Environment, req.RunnerName, req.DindName)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Str("runner", req.RunnerName).Msg("error removing runner buildkit")
	}

	return nil
}

// waitRunnerRedeployJob holds the enable flow back until the runner container
// is running and carries its DinD's label: a redeploy only schedules the
// recreation, and BuildKit has to be deployed on the DinD the recreated runner
// uses, with the runner's config.toml the recreated container reads.
type waitRunnerRedeployJob struct {
	runtimes container_runtime.RuntimeResolver

	target runnerBuildkitTarget
}

func (j *waitRunnerRedeployJob) Do(ctx context.Context) error {
	facts, runtime, err := resolveBuildkitRuntime(ctx, j.runtimes, j.target)
	if err != nil {
		return err
	}

	pollCtx, cancel := context.WithTimeout(ctx, runnerRedeployWait)
	defer cancel()

	ticker := time.NewTicker(runnerRedeployPollPeriod)
	defer ticker.Stop()

	for {
		err = checkRunnerOnDind(pollCtx, runtime, facts)
		if err == nil {
			return nil
		}

		select {
		case <-pollCtx.Done():
			return rerrors.Wrap(err, "timed out waiting for the runner to be redeployed on its dind")
		case <-ticker.C:
		}
	}
}

func checkRunnerOnDind(
	ctx context.Context, runtime container_runtime.ContainerRuntime, facts runnerBuildkitFacts,
) error {
	info, isFound, err := runtime.Inspect(ctx, facts.name)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting runner container")
	}

	if !isFound {
		return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound, facts.name)
	}

	isOnDind := info.Config != nil && info.Config.Labels[labels.RunnerDindLabel] == facts.dindName
	isRunning := info.State != nil && info.State.Running

	if !isOnDind || !isRunning {
		return rerrors.Wrap(errRunnerNotRedeployed, facts.name)
	}

	return nil
}
