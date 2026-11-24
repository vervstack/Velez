package pgaas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

func (s *PostgresServiceImpl) EnablePgReplication(ctx context.Context, name string) error {
	if name == "" {
		return rerrors.Wrap(user_errors.ErrInvalidInstanceName)
	}

	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error getting pg instance service")
	}

	_, err = s.dataStorage.PgInstances().GetPgInstanceByServiceID(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error getting pg instance row")
	}

	payload := &velez_api.EnablePgReplicationTaskPayload{InstanceName: name}

	_, err = s.jobsEngine.EnqueueReplacing(ctx, name, jobs.EnablePgReplicationAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing enable pg replication task")
	}

	return nil
}
