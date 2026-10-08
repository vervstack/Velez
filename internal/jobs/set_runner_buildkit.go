package jobs

import (
	"context"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	SetRunnerBuildkitAction = "set_runner_buildkit"

	stepEnsureBuildkitNetwork    = "ensure_buildkit_network"
	stepInstallBinfmt            = "install_binfmt"
	stepDeployBuildkit           = "deploy_buildkit"
	stepWaitForBuildkitDeploy    = "wait_for_buildkit_deploy"
	stepRegisterBuildkitBinding  = "register_buildkit_binding"
	stepSetBuildkitNetworkMode   = "set_buildkit_network_mode"
	stepClearBuildkitNetworkMode = "clear_buildkit_network_mode"
	stepDropBuildkitSidecar      = "drop_buildkit_sidecar"
	stepRemoveBuildkitNetwork    = "remove_buildkit_network"
	stepRemoveBuildkitVolume     = "remove_buildkit_volume"
)

// runnerBuildkitFacts are the facts about a runner every BuildKit step needs.
type runnerBuildkitFacts struct {
	// name is the runner's service and container name.
	name        string
	environment string
	provider    velez_api.RunnerProvider
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

	facts := runnerBuildkitFacts{
		name:        t.name,
		environment: svc.Env,
		provider:    velez_api.RunnerProvider(velez_api.RunnerProvider_value[runner.Provider]),
	}

	return facts, nil
}

// runnerBuildkitJobs builds the step lists that add and remove a runner's
// BuildKit sidecar - shared by set_runner_buildkit and create_runner.
type runnerBuildkitJobs struct {
	dataStorage storage.Storage
	runtimes    container_runtime.RuntimeResolver
}

// enable puts a BuildKit sidecar on a private network next to the runner and
// points the runner's job containers at that network. Every step is safe to
// repeat.
func (b runnerBuildkitJobs) enable(target runnerBuildkitTarget) []NamedJob {
	return []NamedJob{
		{
			Name: stepEnsureBuildkitNetwork,
			Job:  &ensureBuildkitNetworkJob{runtimes: b.runtimes, target: target},
		},
		{
			Name: stepInstallBinfmt,
			Job: &installBinfmtJob{
				boxes:    b.dataStorage.ResourceBoxes(),
				runtimes: b.runtimes,
				target:   target,
			},
		},
		{
			Name: stepDeployBuildkit,
			Job: &deployBuildkitJob{
				boxes:    b.dataStorage.ResourceBoxes(),
				settings: b.dataStorage.Settings(),
				runtimes: b.runtimes,
				target:   target,
			},
		},
		{
			Name: stepWaitForBuildkitDeploy,
			Job:  &waitBuildkitJob{runtimes: b.runtimes, target: target},
		},
		{
			Name: stepRegisterBuildkitBinding,
			Job:  &registerBuildkitBindingJob{dataStorage: b.dataStorage, target: target},
		},
		{
			Name: stepSetBuildkitNetworkMode,
			Job:  &setBuildkitNetworkModeJob{runtimes: b.runtimes, target: target, isEnabled: true},
		},
	}
}

// disable points the runner's job containers away from the BuildKit network
// first, then drops everything enable created.
func (b runnerBuildkitJobs) disable(target runnerBuildkitTarget) []NamedJob {
	clearNetworkMode := NamedJob{
		Name: stepClearBuildkitNetworkMode,
		Job:  &setBuildkitNetworkModeJob{runtimes: b.runtimes, target: target, isEnabled: false},
	}

	return append([]NamedJob{clearNetworkMode}, b.drop(target)...)
}

// drop removes the sidecar, its binding, the network and the state volume.
func (b runnerBuildkitJobs) drop(target runnerBuildkitTarget) []NamedJob {
	return append(b.dropSidecar(target), b.dropResources(target)...)
}

// dropSidecar removes the sidecar container and its binding.
func (b runnerBuildkitJobs) dropSidecar(target runnerBuildkitTarget) []NamedJob {
	return []NamedJob{
		{
			Name: stepDropBuildkitSidecar,
			Job:  &dropBuildkitSidecarJob{dataStorage: b.dataStorage, runtimes: b.runtimes, target: target},
		},
	}
}

// dropResources removes the network and the state volume, which only go once
// nothing is attached to them any more.
func (b runnerBuildkitJobs) dropResources(target runnerBuildkitTarget) []NamedJob {
	return []NamedJob{
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

// DropRunnerBuildkitSidecar removes a runner's BuildKit sidecar and its binding,
// for a runner that is about to be dropped. No sidecar is not an error.
func DropRunnerBuildkitSidecar(
	ctx context.Context,
	dataStorage storage.Storage,
	runtimes container_runtime.RuntimeResolver,
	name, environment string,
) error {
	steps := runnerBuildkitJobs{dataStorage: dataStorage, runtimes: runtimes}
	facts := runnerBuildkitFacts{name: name, environment: environment}

	for _, step := range steps.dropSidecar(facts) {
		err := step.Job.Do(ctx)
		if err != nil {
			return rerrors.Wrap(err, step.Name)
		}
	}

	return nil
}

// DropRunnerBuildkitResources removes a dropped runner's BuildKit network and
// state volume. Best-effort, like the other cleanup after a service is gone: a
// job container of the removed runner may still hold the network.
func DropRunnerBuildkitResources(
	ctx context.Context,
	dataStorage storage.Storage,
	runtimes container_runtime.RuntimeResolver,
	name, environment string,
) {
	steps := runnerBuildkitJobs{dataStorage: dataStorage, runtimes: runtimes}
	facts := runnerBuildkitFacts{name: name, environment: environment}

	for _, step := range steps.dropResources(facts) {
		err := step.Job.Do(ctx)
		if err != nil {
			log.Ctx(ctx).Warn().
				Str("runner", name).
				Str("step", step.Name).
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
