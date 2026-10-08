package provisioning

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
)

const (
	failedTaskRetention = 5 * time.Minute
)

type Service struct {
	jobsEngine jobs.Engine
}

func New(jobsEngine jobs.Engine) *Service {
	return &Service{
		jobsEngine: jobsEngine,
	}
}

func (s *Service) List(ctx context.Context, actions ...string) ([]domain.ProvisioningTask, error) {
	failedSince := time.Now().Add(-failedTaskRetention)

	entries, err := s.jobsEngine.ListProvisioning(ctx, actions, failedSince)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing provisioning tasks")
	}

	out := make([]domain.ProvisioningTask, 0, len(entries))
	for _, entry := range entries {
		out = append(out, toDomainTask(entry))
	}

	return out, nil
}

func (s *Service) Dismiss(ctx context.Context, entityId, action string) error {
	err := s.jobsEngine.DismissFailed(ctx, entityId, action)
	if err != nil {
		return rerrors.Wrap(err, "error dismissing failed task")
	}

	return nil
}
