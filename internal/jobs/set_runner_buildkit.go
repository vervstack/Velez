package jobs

import (
	"context"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	SetRunnerBuildkitAction = "set_runner_buildkit"

	stepInstallBinfmt            = "install_binfmt"
	stepEnsureBuildkitNetwork    = "ensure_buildkit_network"
	stepDeployBuildkit           = "deploy_buildkit"
	stepWaitForBuildkit          = "wait_for_buildkit"
	stepSetBuildkitNetworkMode   = "set_buildkit_network_mode"
	stepMarkBuildkitEnabled      = "mark_buildkit_enabled"
	stepMarkBuildkitDisabled     = "mark_buildkit_disabled"
	stepClearBuildkitNetworkMode = "clear_buildkit_network_mode"
	stepRemoveBuildkit           = "remove_buildkit"
	stepRemoveBuildkitNetwork    = "remove_buildkit_network"
	stepRemoveBuildkitVolume     = "remove_buildkit_volume"

	// allServicesLimit is a generous ceiling for listing every service
	// when resolving a DinD service id back to its name; the storage clamps it.
	allServicesLimit = 1_000_000
)

// runnerBuildkitFacts are the facts about a runner every BuildKit step needs.
type runnerBuildkitFacts struct {
	// name is the runner's service and container name.
	name        string
	environment string
	provider    velez_api.RunnerProvider
	// dindName is the DinD service the runner uses; empty when it has none.
	dindName string
}

// runnerBuildkitTarget names the runner a BuildKit step works on. The facts are
// read when the step runs: create_runner builds its steps before the runner
// exists.
type runnerBuildkitTarget interface {
	Locate(ctx context.Context) (runnerBuildkitFacts, error)
}

// Locate makes known facts a target of their own.
func (f runnerBuildkitFacts) Locate(_ context.Context) (runnerBuildkitFacts, error) {
	return f, nil
}

// createdRunnerTarget is the runner a create_runner task is creating.
type createdRunnerTarget struct {
	req          createRunnerRequestAccessor
	instanceName string
}

func (t createdRunnerTarget) Locate(_ context.Context) (runnerBuildkitFacts, error) {
	request := t.req.GetRequest()
	provider, _, _ := runnerProviderConfig(request)

	facts := runnerBuildkitFacts{
		name:        t.instanceName,
		environment: request.GetEnvironment(),
		provider:    provider,
		dindName:    request.GetDindName(),
	}

	return facts, nil
}

// storedRunnerTarget is a runner that already exists, read from storage.
type storedRunnerTarget struct {
	services storage.ServicesStorage
	runners  storage.RunnersStorage

	name string
}

func (t storedRunnerTarget) Locate(ctx context.Context) (runnerBuildkitFacts, error) {
	svc, err := t.services.GetByName(ctx, t.name)
	if err != nil {
		return runnerBuildkitFacts{}, rerrors.Wrap(err, "error getting runner service")
	}

	runner, err := t.runners.GetRunnerByServiceID(ctx, svc.ID)
	if err != nil {
		return runnerBuildkitFacts{}, rerrors.Wrap(err, "error getting runner row")
	}

	dindName, err := dindNameByServiceId(ctx, t.services, runner.DindServiceId)
	if err != nil {
		return runnerBuildkitFacts{}, err
	}

	facts := runnerBuildkitFacts{
		name:        t.name,
		environment: svc.Env,
		provider:    velez_api.RunnerProvider(velez_api.RunnerProvider_value[runner.Provider]),
		dindName:    dindName,
	}

	return facts, nil
}

// dindNameByServiceId resolves a DinD service id back to its name; 0 means the
// runner has no DinD. storage.ServicesStorage has no id lookup, so every service
// is resolved by name, as runneraas does for its runner list.
func dindNameByServiceId(ctx context.Context, services storage.ServicesStorage, dindServiceId int64) (string, error) {
	if dindServiceId == 0 {
		return "", nil
	}

	listReq := domain.ListServicesReq{
		IncludeInternal: true,
		Paging:          domain.Paging{Limit: allServicesLimit},
	}

	list, err := services.List(ctx, listReq)
	if err != nil {
		return "", rerrors.Wrap(err, "error listing services")
	}

	for _, base := range list.Services {
		svc, getErr := services.GetByName(ctx, base.Name)
		if getErr != nil {
			continue
		}

		if svc.ID == dindServiceId {
			return base.Name, nil
		}
	}

	return "", rerrors.Wrap(user_errors.ErrDindNotFound)
}

// runnerBuildkitJobs builds the step lists that add and remove a runner's
// buildkitd inside its DinD - shared by set_runner_buildkit and create_runner.
type runnerBuildkitJobs struct {
	dataStorage storage.Storage
	runtimes    container_runtime.RuntimeResolver
}

// enable runs buildkitd inside the runner's DinD on a private network and points
// the runner's job containers at that network. Every step is safe to repeat.
func (b runnerBuildkitJobs) enable(target runnerBuildkitTarget) []NamedJob {
	return []NamedJob{
		{
			Name: stepInstallBinfmt,
			Job: &installBinfmtJob{
				boxes:    b.dataStorage.ResourceBoxes(),
				runtimes: b.runtimes,
				target:   target,
			},
		},
		{
			Name: stepEnsureBuildkitNetwork,
			Job:  &ensureBuildkitNetworkJob{runtimes: b.runtimes, target: target},
		},
		{
			Name: stepDeployBuildkit,
			Job: &deployBuildkitJob{
				boxes:    b.dataStorage.ResourceBoxes(),
				runtimes: b.runtimes,
				target:   target,
			},
		},
		{
			Name: stepWaitForBuildkit,
			Job:  &waitBuildkitJob{runtimes: b.runtimes, target: target},
		},
		{
			Name: stepSetBuildkitNetworkMode,
			Job:  &setBuildkitNetworkModeJob{runtimes: b.runtimes, target: target, isEnabled: true},
		},
		{
			Name: stepMarkBuildkitEnabled,
			Job:  b.mark(target, true),
		},
	}
}

// disable records BuildKit as off and points the runner's job containers away
// from its network first, then drops everything enable created.
func (b runnerBuildkitJobs) disable(target runnerBuildkitTarget) []NamedJob {
	return []NamedJob{
		{
			Name: stepMarkBuildkitDisabled,
			Job:  b.mark(target, false),
		},
		{
			Name: stepClearBuildkitNetworkMode,
			Job:  &setBuildkitNetworkModeJob{runtimes: b.runtimes, target: target, isEnabled: false},
		},
		{
			Name: stepRemoveBuildkit,
			Job:  &removeBuildkitJob{runtimes: b.runtimes, target: target},
		},
		{
			Name: stepRemoveBuildkitNetwork,
			Job:  &removeBuildkitNetworkJob{runtimes: b.runtimes, target: target},
		},
		{
			Name: stepRemoveBuildkitVolume,
			Job:  &removeBuildkitVolumeJob{runtimes: b.runtimes, target: target},
		},
	}
}

func (b runnerBuildkitJobs) mark(target runnerBuildkitTarget, isEnabled bool) *markBuildkitJob {
	return &markBuildkitJob{
		services:  b.dataStorage.Services(),
		runners:   b.dataStorage.Runners(),
		target:    target,
		isEnabled: isEnabled,
	}
}

// DropRunnerBuildkit removes a dropped runner's buildkitd, network and state
// volume from its DinD. Best-effort, like the other cleanup after a service is
// gone: the DinD may be gone too, or a job container may still hold the network.
func DropRunnerBuildkit(
	ctx context.Context,
	runtimes container_runtime.RuntimeResolver,
	name, environment, dindName string,
) {
	dind, closer, err := runtimes.NestedRuntime(ctx, environment, dindName)
	if err != nil {
		log.Ctx(ctx).Warn().
			Str("runner", name).
			Str("dind", dindName).
			Err(err).
			Msg("error opening dind daemon to remove runner buildkit")

		return
	}

	defer common.CloseWithLog(closer.Close, "buildkit dind client")

	steps := []struct {
		name string
		run  func(context.Context, container_runtime.ContainerRuntime, string) error
	}{
		{stepRemoveBuildkit, removeBuildkitd},
		{stepRemoveBuildkitNetwork, removeBuildkitNetwork},
		{stepRemoveBuildkitVolume, removeBuildkitVolume},
	}

	for _, step := range steps {
		err = step.run(ctx, dind, name)
		if err != nil {
			log.Ctx(ctx).Warn().
				Str("runner", name).
				Str("step", step.name).
				Err(err).
				Msg("error removing runner buildkit resource")
		}
	}
}

type setRunnerBuildkitHandler struct {
	steps runnerBuildkitJobs

	dataStorage storage.Storage
}

func NewSetRunnerBuildkitHandler(
	dataStorage storage.Storage,
	runtimes container_runtime.RuntimeResolver,
) TaskHandler {
	return &setRunnerBuildkitHandler{
		steps:       runnerBuildkitJobs{dataStorage: dataStorage, runtimes: runtimes},
		dataStorage: dataStorage,
	}
}

func (h *setRunnerBuildkitHandler) Action() string {
	return SetRunnerBuildkitAction
}

func (h *setRunnerBuildkitHandler) NewContext() TaskContext {
	return &velez_api.SetRunnerBuildkitTaskPayload{}
}

func (h *setRunnerBuildkitHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.SetRunnerBuildkitTaskPayload)
	if !ok {
		panic("set_runner_buildkit: BuildJobs called with mismatched TaskContext type")
	}

	request := payload.GetRequest()

	target := storedRunnerTarget{
		services: h.dataStorage.Services(),
		runners:  h.dataStorage.Runners(),
		name:     request.GetName(),
	}

	if request.GetIsBuildkitEnabled() {
		return h.steps.enable(target)
	}

	return h.steps.disable(target)
}
