package runneraas

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (s *RunneraasService) DropRunner(ctx context.Context, name string) error {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error getting runner service")
	}

	runner, err := s.dataStorage.Runners().GetRunnerByServiceID(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error getting runner row")
	}

	dindName := s.runnerDindName(ctx, name, runner)

	removeReq := domain.RemoveServiceReq{Name: name, DropRunningInstances: true}

	err = s.vervServices.Remove(ctx, removeReq)
	if err != nil {
		return rerrors.Wrap(err, "error removing runner service")
	}

	if dindName != "" {
		jobs.DropRunnerBuildkit(ctx, s.runtimes, name, svc.Env, dindName)
	}

	err = s.dataStorage.Runners().DeleteRunner(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error deleting runner row")
	}

	secretRef, err := domain.ParseSecretRef(runner.SecretRef)
	if err != nil {
		return rerrors.Wrap(err, "error parsing runner secret ref")
	}

	err = s.secrets.Delete(ctx, secretRef)
	if err != nil && !errors.Is(err, user_errors.ErrSecretNotFound) {
		return rerrors.Wrap(err, "error deleting runner secret")
	}

	return nil
}

// runnerDindName is the DinD a runner with BuildKit keeps buildkitd in; empty for
// a runner without BuildKit or whose DinD can no longer be resolved, so dropping
// the runner never depends on reaching a DinD.
func (s *RunneraasService) runnerDindName(ctx context.Context, name string, runner domain.Runner) string {
	if !runner.IsBuildkitEnabled || runner.DindServiceId == 0 {
		return ""
	}

	dindName, err := s.dindName(ctx, runner.DindServiceId)
	if err != nil {
		log.Ctx(ctx).Warn().
			Str("runner", name).
			Err(err).
			Msg("error resolving dind of runner, skipping buildkit cleanup")

		return ""
	}

	return dindName
}
