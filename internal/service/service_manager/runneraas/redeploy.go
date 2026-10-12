package runneraas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
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

	overlay, err := s.runnerOverlay(ctx, runner)
	if err != nil {
		return rerrors.Wrap(err, "error building runner overlay")
	}

	upgradeReq := domain.UpgradeDeployReq{
		ServiceName:   name,
		EnvOverrides:  overlay.env,
		ExtraNetworks: overlay.networks,
	}

	err = s.vervServices.UpgradeDeploy(ctx, upgradeReq)
	if err != nil {
		return rerrors.Wrap(err, "error redeploying runner")
	}

	return nil
}

// runnerOverlay is what a runner's recreated container must carry beyond its
// stored spec: its DOCKER_HOST and, on a DinD, the DinD's network. dindName is
// empty for a runner without a DinD.
type runnerOverlay struct {
	env      map[string]string
	networks []*velez_api.NetworkBind
	dindName string
}

func (s *RunneraasService) runnerOverlay(ctx context.Context, runner domain.Runner) (runnerOverlay, error) {
	overlay := runnerOverlay{
		env: map[string]string{redeployDockerHostEnvVar: runner.DockerSocketAddress},
	}

	if runner.DindServiceId == 0 {
		return overlay, nil
	}

	dindName, err := s.dindName(ctx, runner.DindServiceId)
	if err != nil {
		return runnerOverlay{}, rerrors.Wrap(err, "error resolving runner dind")
	}

	dindNetwork := &velez_api.NetworkBind{NetworkName: domain.DindNetworkName(dindName)}

	overlay.env[redeployDockerHostEnvVar] = domain.DindAddress(dindName)
	overlay.networks = []*velez_api.NetworkBind{dindNetwork}
	overlay.dindName = dindName

	return overlay, nil
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
