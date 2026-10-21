// Package settings implements service.SettingsService - node-wide Velez
// settings.
package settings

import (
	"context"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/clients/node_clients/runtime_policy"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	sysboxRuntimeName = "sysbox-runc"
)

type SettingsServiceImpl struct {
	dataStorage storage.Storage
	docker      node_clients.Docker
}

func New(dataStorage storage.Storage, docker node_clients.Docker) *SettingsServiceImpl {
	return &SettingsServiceImpl{
		dataStorage: dataStorage,
		docker:      docker,
	}
}

func (s *SettingsServiceImpl) GetSettings(ctx context.Context) (domain.Settings, error) {
	settings, err := s.dataStorage.Settings().GetSettings(ctx)
	if err != nil {
		return domain.Settings{}, rerrors.Wrap(err, "error getting settings")
	}

	return settings, nil
}

func (s *SettingsServiceImpl) UpdateSettings(
	ctx context.Context, req domain.UpdateSettingsReq,
) (domain.Settings, error) {
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

func (s *SettingsServiceImpl) ListWhitelistedImages(ctx context.Context) ([]domain.WhitelistedImageEntry, error) {
	stored, err := s.dataStorage.Whitelist().List(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing whitelisted images")
	}

	return buildWhitelistEntries(runtime_policy.BuiltinWhitelist(), stored), nil
}

func (s *SettingsServiceImpl) RemoveWhitelistedImage(ctx context.Context, id string) error {
	imageUuid, err := uuid.Parse(id)
	if err != nil {
		return rerrors.Wrap(user_errors.ErrWhitelistedImageIdInvalid, id)
	}

	err = s.dataStorage.Whitelist().Remove(ctx, imageUuid)
	if err != nil {
		return rerrors.Wrap(err, "error removing whitelisted image")
	}

	return nil
}

func buildWhitelistEntries(
	builtin []string, stored []domain.WhitelistedImage,
) []domain.WhitelistedImageEntry {
	entries := make([]domain.WhitelistedImageEntry, 0, len(builtin)+len(stored))

	for _, repository := range builtin {
		entry := domain.WhitelistedImageEntry{
			Repository: repository,
			IsBuiltin:  true,
		}

		entries = append(entries, entry)
	}

	for _, image := range stored {
		entry := domain.WhitelistedImageEntry{
			Id:         image.Id.String(),
			Repository: image.Repository,
		}

		entries = append(entries, entry)
	}

	return entries
}

func (s *SettingsServiceImpl) ensureSysboxRuntime(ctx context.Context) error {
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
