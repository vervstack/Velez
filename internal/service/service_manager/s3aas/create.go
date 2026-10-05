package s3aas

import (
	"context"
	"regexp"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	supportedReplicationFactor = 1
	maxInstanceNameLength      = 63
)

var instanceNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

func (s *Service) CreateInstance(ctx context.Context, req *velez_api.CreateS3Instance_Request) error {
	err := validateCreateRequest(req)
	if err != nil {
		return rerrors.Wrap(err)
	}

	for _, name := range []string{domain.S3ServiceName(req.GetName()), domain.S3WebUiServiceName(req.GetName())} {
		err = s.ensureNameIsFree(ctx, name)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	initialContext := &velez_api.CreateS3InstanceTaskPayload{
		Request: req,
	}

	_, err = s.jobsEngine.Enqueue(ctx, req.GetName(), jobs.CreateS3InstanceAction, initialContext)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing create s3 instance task")
	}

	return nil
}

func validateCreateRequest(req *velez_api.CreateS3Instance_Request) error {
	if req.GetName() == "" {
		return rerrors.Wrap(user_errors.ErrS3InstanceNameRequired)
	}

	if len(req.GetName()) > maxInstanceNameLength || !instanceNamePattern.MatchString(req.GetName()) {
		return rerrors.Wrap(user_errors.ErrS3InstanceNameInvalid)
	}

	if req.ReplicationFactor != nil && req.GetReplicationFactor() != supportedReplicationFactor {
		return rerrors.Wrap(user_errors.ErrS3ReplicationFactorUnsupported)
	}

	return nil
}

func (s *Service) ensureNameIsFree(ctx context.Context, name string) error {
	_, err := s.dataStorage.Services().GetByName(ctx, name)
	if err == nil {
		return rerrors.Wrap(user_errors.ErrS3InstanceAlreadyExists)
	}

	if !rerrors.Is(err, user_errors.ErrStorageNotFound) {
		return rerrors.Wrap(err, "error checking service name")
	}

	return nil
}
