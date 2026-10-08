package dinds

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (s *Service) DropDind(ctx context.Context, name string) error {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrStorageNotFound) {
			return rerrors.Wrap(user_errors.ErrDindNotFound)
		}

		return rerrors.Wrap(err, "error getting dind service")
	}

	_, err = s.dataStorage.DindInstances().GetDindInstanceByServiceId(ctx, svc.ID)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrStorageNotFound) {
			return rerrors.Wrap(user_errors.ErrDindNotFound)
		}

		return rerrors.Wrap(err, "error getting dind instance")
	}

	err = s.ensureNoRunnerUsesDind(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err)
	}

	payload := &velez_api.DropDindTaskPayload{Name: name}

	_, err = s.jobsEngine.EnqueueReplacing(ctx, name, jobs.DropDindAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing drop dind task")
	}

	return nil
}

func (s *Service) ensureNoRunnerUsesDind(ctx context.Context, dindServiceId int64) error {
	runners, err := s.dataStorage.Runners().ListRunners(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error listing runners")
	}

	for _, runner := range runners {
		if runner.DindServiceId == dindServiceId {
			return rerrors.Wrap(user_errors.ErrDindInUse)
		}
	}

	return nil
}
