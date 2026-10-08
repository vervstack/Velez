package pgaas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (s *PgaasService) CreatePgInstance(ctx context.Context, req domain.CreatePgInstanceReq) error {
	if req.Isolation == velez_api.PgInstanceIsolation_PG_INSTANCE_ISOLATION_SHARED_POOL {
		return rerrors.Wrap(user_errors.ErrPgSharedPoolIsolationNotSupported)
	}

	request := &velez_api.CreatePgInstance_Request{
		Name:         req.Name,
		Environment:  optionalString(req.Environment),
		Box:          optionalString(req.Box),
		OwnerService: optionalString(req.OwnerService),
	}

	if req.ExposeToPort != 0 {
		request.ExposeToPort = &req.ExposeToPort
	}

	payload := &velez_api.CreatePgInstanceTaskPayload{Request: request}

	_, err := s.jobsEngine.EnqueueReplacing(ctx, req.Name, jobs.CreatePgInstanceAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing create pg instance task")
	}

	return nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
