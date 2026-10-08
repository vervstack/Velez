package pgaas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
)

func (s *PgaasService) DropPgInstance(ctx context.Context, name string) error {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error getting pg instance service")
	}

	_, err = s.dataStorage.PgInstances().GetPgInstanceByServiceID(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error getting pg instance row")
	}

	payload := &velez_api.DropPgInstanceTaskPayload{Name: name}

	_, err = s.jobsEngine.EnqueueReplacing(ctx, name, jobs.DropPgInstanceAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing drop pg instance task")
	}

	return nil
}
