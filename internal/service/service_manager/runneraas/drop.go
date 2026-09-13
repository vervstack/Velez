package runneraas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
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

	removeReq := domain.RemoveServiceReq{Name: name, DropRunningInstances: true}

	err = s.vervServices.Remove(ctx, removeReq)
	if err != nil {
		return rerrors.Wrap(err, "error removing runner service")
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
	if err != nil {
		return rerrors.Wrap(err, "error deleting runner secret")
	}

	return nil
}
