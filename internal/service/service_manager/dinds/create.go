package dinds

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	sysboxRuntimeName = "sysbox-runc"
)

func (s *Service) CreateDind(ctx context.Context, req domain.CreateDindReq) error {
	err := domain.ValidateInstanceName(req.Name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = validateDindName(req.Name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = s.ensureNameIsFree(ctx, req.Name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if req.IsSysboxEnabled {
		err = s.ensureSysboxRuntime(ctx)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	initialContext := &velez_api.CreateDindTaskPayload{
		Request: dindRequestToPb(req),
	}

	_, err = s.jobsEngine.EnqueueReplacing(ctx, req.Name, jobs.CreateDindAction, initialContext)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing create dind task")
	}

	return nil
}

func (s *Service) ensureNameIsFree(ctx context.Context, name string) error {
	_, err := s.dataStorage.Services().GetByName(ctx, name)
	if err == nil {
		return rerrors.Wrap(user_errors.ErrDindNameTaken)
	}

	if !rerrors.Is(err, user_errors.ErrStorageNotFound) {
		return rerrors.Wrap(err, "error checking service name")
	}

	return nil
}

func (s *Service) ensureSysboxRuntime(ctx context.Context) error {
	isRegistered, err := dockerutils.HasRuntime(ctx, s.docker.Client(), sysboxRuntimeName)
	if err != nil {
		return rerrors.Wrap(err, "error checking sysbox runtime")
	}

	if !isRegistered {
		return rerrors.Wrap(user_errors.ErrSysboxRuntimeUnavailable)
	}

	return nil
}

func dindRequestToPb(req domain.CreateDindReq) *velez_api.CreateDind_Request {
	pbReq := &velez_api.CreateDind_Request{
		Name:            req.Name,
		IsSysboxEnabled: &req.IsSysboxEnabled,
	}

	if req.Environment != "" {
		pbReq.Environment = &req.Environment
	}

	return pbReq
}
