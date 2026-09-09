package pgaas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (s *PgaasService) DropPgInstance(ctx context.Context, name string) error {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error getting pg instance service")
	}

	instance, err := s.dataStorage.PgInstances().GetPgInstanceByServiceID(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error getting pg instance row")
	}

	removeReq := domain.RemoveServiceReq{Name: name, DropRunningInstances: true}

	err = s.vervServices.Remove(ctx, removeReq)
	if err != nil {
		return rerrors.Wrap(err, "error removing pg instance service")
	}

	err = s.dataStorage.PgInstances().DeletePgInstance(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error deleting pg instance row")
	}

	secretRef, err := domain.ParseSecretRef(instance.SecretRef)
	if err != nil {
		return rerrors.Wrap(err, "error parsing pg instance secret ref")
	}

	err = s.secrets.Delete(ctx, secretRef)
	if err != nil {
		return rerrors.Wrap(err, "error deleting pg instance secret")
	}

	return nil
}
