package dinds

import (
	"context"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
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

	removeReq := domain.RemoveServiceReq{Name: name, DropRunningInstances: true}

	err = s.vervServices.Remove(ctx, removeReq)
	if err != nil {
		return rerrors.Wrap(err, "error removing dind service")
	}

	err = s.dataStorage.DindInstances().DeleteDindInstance(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error deleting dind instance row")
	}

	err = s.docker.Client().NetworkRemove(ctx, domain.DindNetworkName(name))
	if err != nil && !cerrdefs.IsNotFound(err) {
		log.Ctx(ctx).Warn().
			Str("dind_name", name).
			Err(err).
			Msg("error removing dind network")
	}

	err = s.docker.Client().VolumeRemove(ctx, domain.DindDataVolumeName(name), false)
	if err != nil && !cerrdefs.IsNotFound(err) {
		log.Ctx(ctx).Warn().
			Str("dind_name", name).
			Err(err).
			Msg("error removing dind data volume")
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
