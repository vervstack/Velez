package vcn

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	dropTimeout = time.Minute
)

func (s *Service) Drop(ctx context.Context, id string) error {
	record, err := s.find(ctx, id)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if record.Source != sourceDeployed || record.NodeName == "" {
		return s.drop(ctx, record)
	}

	current, err := s.CurrentNode(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if current == record.NodeName {
		return s.drop(ctx, record)
	}

	payload := &jobs.DropVcnNetworkPayload{NetworkId: id, NodeName: record.NodeName}

	task, err := s.engine.Enqueue(ctx, id, jobs.DropVcnNetworkAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing VPN network removal")
	}

	watchCtx, cancel := context.WithTimeout(ctx, dropTimeout)
	defer cancel()

	for update := range s.engine.WatchAfter(watchCtx, id, jobs.DropVcnNetworkAction, task.ID-1) {
		if update.Status == tasks_queries.VelezTaskStatusDONE {
			return nil
		}

		if update.Status == tasks_queries.VelezTaskStatusFAILED {
			return rerrors.Wrap(user_errors.ErrTaskFailed, update.Error.String)
		}
	}

	return user_errors.New("timed out waiting for the VPN network's node to remove it")
}

func (s *Service) DropLocal(ctx context.Context, id string) error {
	record, err := s.find(ctx, id)
	if rerrors.Is(err, user_errors.ErrVcnNetworkNotFound) {
		return nil
	}

	if err != nil {
		return rerrors.Wrap(err)
	}

	current, err := s.CurrentNode(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if record.NodeName != current {
		return rerrors.Wrap(user_errors.ErrVcnNetworkOwnerMismatch)
	}

	return s.drop(ctx, record)
}

func (s *Service) drop(ctx context.Context, record domain.Network) error {
	repo := s.storage.Networks()

	if record.Source == sourceDeployed {
		err := s.removeDeployed(ctx, record)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	if repo == nil {
		return nil
	}

	err := repo.Drop(ctx, record.Id)
	if err != nil && !rerrors.Is(err, user_errors.ErrStorageNotFound) {
		return rerrors.Wrap(err)
	}

	return nil
}

func (s *Service) removeDeployed(ctx context.Context, record domain.Network) error {
	runtime, err := s.runtimes.Runtime(ctx, "")
	if err != nil {
		return rerrors.Wrap(err)
	}

	info, found, err := runtime.Inspect(ctx, record.ContainerId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if found && (info.Config == nil || info.Config.Labels[networkIdLabel] != record.Id) {
		return rerrors.Wrap(user_errors.ErrVcnNetworkOwnerMismatch)
	}

	err = runtime.Remove(ctx, record.ContainerId)
	if err != nil {
		return rerrors.Wrap(err, "error removing Headscale container")
	}

	err = runtime.RemoveVolume(ctx, deployedVolume(record))
	if err != nil {
		return rerrors.Wrap(err, "error removing Headscale data volume")
	}

	if found {
		s.releasePort(record.Url)
	}

	if s.storage.Networks() == nil {
		return nil
	}

	err = s.storage.Services().Delete(ctx, record.ServiceName)
	if err != nil && !rerrors.Is(err, user_errors.ErrStorageNotFound) {
		return rerrors.Wrap(err, "error removing Headscale service registration")
	}

	return nil
}

func (s *Service) releasePort(endpoint string) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return
	}

	port, err := strconv.ParseUint(parsed.Port(), 10, 16)
	if err != nil {
		return
	}

	s.node.PortManager().UnlockPorts([]uint32{uint32(port)})
}
