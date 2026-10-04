// Package settings implements service.SettingsService - node-wide Velez
// settings.
package settings

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	sysboxRuntimeName = "sysbox-runc"
)

type Service struct {
	dataStorage storage.Storage
	docker      node_clients.Docker
}

func New(dataStorage storage.Storage, docker node_clients.Docker) *Service {
	return &Service{
		dataStorage: dataStorage,
		docker:      docker,
	}
}

func (s *Service) GetSettings(ctx context.Context) (domain.Settings, error) {
	settings, err := s.dataStorage.Settings().GetSettings(ctx)
	if err != nil {
		return domain.Settings{}, rerrors.Wrap(err, "error getting settings")
	}

	return settings, nil
}

func (s *Service) UpdateSettings(ctx context.Context, req domain.UpdateSettingsReq) (domain.Settings, error) {
	current, err := s.dataStorage.Settings().GetSettings(ctx)
	if err != nil {
		return domain.Settings{}, rerrors.Wrap(err, "error getting settings")
	}

	updated := mergeSettings(current, req)

	if updated.IsSysboxEnabled && !current.IsSysboxEnabled {
		err = s.ensureSysboxRuntime(ctx)
		if err != nil {
			return domain.Settings{}, rerrors.Wrap(err)
		}
	}

	result, err := s.dataStorage.Settings().UpdateSettings(ctx, updated)
	if err != nil {
		return domain.Settings{}, rerrors.Wrap(err, "error updating settings")
	}

	return result, nil
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

func mergeSettings(current domain.Settings, req domain.UpdateSettingsReq) domain.Settings {
	if req.IsSysboxEnabled != nil {
		current.IsSysboxEnabled = *req.IsSysboxEnabled
	}

	if req.IsSysboxWhitelistIgnored != nil {
		current.IsSysboxWhitelistIgnored = *req.IsSysboxWhitelistIgnored
	}

	return current
}
