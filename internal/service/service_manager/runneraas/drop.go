package runneraas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
)

func (s *RunneraasService) DropRunner(ctx context.Context, name string) error {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error getting runner service")
	}

	_, err = s.dataStorage.Runners().GetRunnerByServiceID(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error getting runner row")
	}

	payload := &velez_api.DropRunnerTaskPayload{Name: name}

	_, err = s.jobsEngine.EnqueueReplacing(ctx, name, jobs.DropRunnerAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing drop runner task")
	}

	return nil
}
