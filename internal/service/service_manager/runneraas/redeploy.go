package runneraas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	// redeployDockerHostEnvVar mirrors internal/jobs/create_runner.go's
	// runnerDockerHostEnvVar - the same env var, set at redeploy time
	// instead of create time. Not imported from there: jobs depends on
	// service, not the other way around.
	redeployDockerHostEnvVar = "DOCKER_HOST"
)

// RedeployRunner recreates the runner's container from its current stored
// deployment spec, overlaying the runner row's stored docker_socket_address
// as DOCKER_HOST - see runners_api.proto's RedeployRunner doc comment on why
// a plain restart/reregister can't apply this instead. A runner on a DinD
// keeps both its DOCKER_HOST and its attachment to the DinD's network.
func (s *RunneraasService) RedeployRunner(ctx context.Context, name string) error {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error getting runner service")
	}

	runner, err := s.dataStorage.Runners().GetRunnerByServiceID(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error getting runner row")
	}

	overlay, err := s.runnerOverlay(ctx, svc.Env, name, runner)
	if err != nil {
		return rerrors.Wrap(err, "error building runner overlay")
	}

	upgradeReq := domain.UpgradeDeployReq{
		ServiceName:    name,
		EnvOverrides:   overlay.env,
		ExtraNetworks:  overlay.networks,
		RemoveNetworks: overlay.removedNetworks,
		LabelOverrides: overlay.labels,
	}

	err = s.vervServices.UpgradeDeploy(ctx, upgradeReq)
	if err != nil {
		return rerrors.Wrap(err, "error redeploying runner")
	}

	if runner.IsBuildkitEnabled && overlay.dindName != "" {
		_, err = s.enqueueSetRunnerBuildkit(ctx, name, runner, true, true)
		if err != nil {
			return rerrors.Wrap(err, "error restoring runner buildkit")
		}
	}

	return nil
}

// runnerOverlay is what a runner's recreated container must carry beyond its
// stored spec: its DOCKER_HOST and DinD label and, on a DinD, the DinD's
// network - and must drop of it, the network of the DinD it leaves. dindName is
// empty for a runner without a DinD.
type runnerOverlay struct {
	env             map[string]string
	labels          map[string]string
	networks        []*velez_api.NetworkBind
	removedNetworks []string
	dindName        string
}

func (s *RunneraasService) runnerOverlay(
	ctx context.Context, environment, name string, runner domain.Runner,
) (runnerOverlay, error) {
	overlay := runnerOverlay{
		env:    map[string]string{redeployDockerHostEnvVar: runner.DockerSocketAddress},
		labels: map[string]string{labels.RunnerDindLabel: ""},
	}

	if runner.DindServiceId != 0 {
		dindName, err := s.dindName(ctx, runner.DindServiceId)
		if err != nil {
			return runnerOverlay{}, rerrors.Wrap(err, "error resolving runner dind")
		}

		dindNetwork := &velez_api.NetworkBind{NetworkName: domain.DindNetworkName(dindName)}

		overlay.env[redeployDockerHostEnvVar] = domain.DindAddress(dindName)
		overlay.labels[labels.RunnerDindLabel] = dindName
		overlay.networks = []*velez_api.NetworkBind{dindNetwork}
		overlay.dindName = dindName
	}

	currentDind, err := s.currentContainerDind(ctx, environment, name)
	if err != nil {
		return runnerOverlay{}, rerrors.Wrap(err)
	}

	if currentDind != "" && currentDind != overlay.dindName {
		overlay.removedNetworks = []string{domain.DindNetworkName(currentDind)}
	}

	return overlay, nil
}

// currentContainerDind is the DinD the running runner container is wired to
// (its label), which can differ from the runner row until a redeploy; empty
// when the container is gone or has none.
func (s *RunneraasService) currentContainerDind(ctx context.Context, environment, name string) (string, error) {
	runtime, err := s.runtimes.Runtime(ctx, environment)
	if err != nil {
		return "", rerrors.Wrap(err, "error resolving container runtime")
	}

	info, isFound, err := runtime.Inspect(ctx, name)
	if err != nil {
		return "", rerrors.Wrap(err, "error inspecting runner container")
	}

	if !isFound || info.Config == nil {
		return "", nil
	}

	return info.Config.Labels[labels.RunnerDindLabel], nil
}

func (s *RunneraasService) dindName(ctx context.Context, dindServiceId int64) (string, error) {
	stateById, err := s.serviceStateByID(ctx)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	base, isFound := stateById[dindServiceId]
	if !isFound {
		return "", rerrors.Wrap(user_errors.ErrDindNotFound)
	}

	return base.Name, nil
}
