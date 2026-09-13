package verv_services

import (
	"context"
	"database/sql"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (v *VervService) ListRegistries(ctx context.Context) ([]domain.Registry, error) {
	list, err := v.registries().ListRegistries(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing registries from storage")
	}

	return list, nil
}

func (v *VervService) GetRegistry(ctx context.Context, id int64) (domain.Registry, error) {
	reg, err := v.registries().GetRegistryByID(ctx, id)
	if err != nil {
		return domain.Registry{}, rerrors.Wrapf(user_errors.ErrRegistryNotFound, "unknown registry id %d", id)
	}

	return reg, nil
}

// CreateRegistry persists a new registry. When req.IsDefault is true, it runs
// inside a transaction that first clears is_default on every other row, so at
// most one registry is ever the default.
func (v *VervService) CreateRegistry(ctx context.Context, req domain.CreateRegistryReq) (domain.Registry, error) {
	if req.Name == "" {
		return domain.Registry{}, rerrors.Wrap(user_errors.ErrRegistryNameRequired)
	}

	if !isValidRegistryType(req.Type) {
		return domain.Registry{}, rerrors.Wrap(user_errors.ErrInvalidRegistryType)
	}

	if !req.IsDefault {
		reg, err := v.registries().CreateRegistry(ctx, req)
		if err != nil {
			return domain.Registry{}, rerrors.Wrap(err, "error creating registry")
		}

		return reg, nil
	}

	var created domain.Registry

	err := v.dataStorage.TxManager().Execute(func(tx *sql.Tx) error {
		repo := v.registries().WithTx(tx)

		clearErr := repo.ClearDefaultRegistry(ctx)
		if clearErr != nil {
			return rerrors.Wrap(clearErr, "error clearing default registry")
		}

		var createErr error

		created, createErr = repo.CreateRegistry(ctx, req)
		if createErr != nil {
			return rerrors.Wrap(createErr, "error creating registry")
		}

		return nil
	})
	if err != nil {
		return domain.Registry{}, rerrors.Wrap(err, "error creating registry")
	}

	return created, nil
}

// UpdateRegistry mirrors CreateRegistry's single-default handling: setting
// is_default = true runs inside a transaction that clears every other row
// first.
func (v *VervService) UpdateRegistry(ctx context.Context, req domain.UpdateRegistryReq) (domain.Registry, error) {
	if req.Id == 0 {
		return domain.Registry{}, user_errors.ErrRegistryIdRequired
	}

	if req.Type != nil && !isValidRegistryType(*req.Type) {
		return domain.Registry{}, rerrors.Wrap(user_errors.ErrInvalidRegistryType)
	}

	if req.IsDefault == nil || !*req.IsDefault {
		reg, err := v.registries().UpdateRegistry(ctx, req)
		if err != nil {
			return domain.Registry{}, rerrors.Wrap(err, "error updating registry")
		}

		return reg, nil
	}

	var updated domain.Registry

	err := v.dataStorage.TxManager().Execute(func(tx *sql.Tx) error {
		repo := v.registries().WithTx(tx)

		clearErr := repo.ClearDefaultRegistry(ctx)
		if clearErr != nil {
			return rerrors.Wrap(clearErr, "error clearing default registry")
		}

		var updateErr error

		updated, updateErr = repo.UpdateRegistry(ctx, req)
		if updateErr != nil {
			return rerrors.Wrap(updateErr, "error updating registry")
		}

		return nil
	})
	if err != nil {
		return domain.Registry{}, rerrors.Wrap(err, "error updating registry")
	}

	return updated, nil
}

// DeleteRegistry removes a registry identified by id or by name. It does not
// auto-promote a new default when the current default is deleted - leaving
// zero defaults is acceptable, a deliberate scope cut.
func (v *VervService) DeleteRegistry(ctx context.Context, req domain.DeleteRegistryReq) error {
	target, err := v.resolveRegistryDeleteTarget(ctx, req)
	if err != nil {
		return err
	}

	err = v.registries().DeleteRegistry(ctx, target.Id)
	if err != nil {
		return rerrors.Wrap(err, "error deleting registry")
	}

	return nil
}

func (v *VervService) resolveRegistryDeleteTarget(
	ctx context.Context,
	req domain.DeleteRegistryReq,
) (domain.Registry, error) {
	switch {
	case req.Id != nil && *req.Id != 0:
		reg, err := v.registries().GetRegistryByID(ctx, *req.Id)
		if err != nil {
			return domain.Registry{}, rerrors.Wrap(user_errors.ErrRegistryNotFound, "unknown registry id")
		}

		return reg, nil
	case req.Name != nil && *req.Name != "":
		list, err := v.registries().ListRegistries(ctx)
		if err != nil {
			return domain.Registry{}, rerrors.Wrap(err, "error listing registries")
		}

		for _, reg := range list {
			if reg.Name == *req.Name {
				return reg, nil
			}
		}

		return domain.Registry{}, rerrors.Wrapf(user_errors.ErrRegistryNotFound, "unknown registry %q", *req.Name)
	default:
		return domain.Registry{}, user_errors.ErrRegistryIdOrNameRequired
	}
}

func isValidRegistryType(t domain.RegistryType) bool {
	switch t {
	case domain.RegistryTypeDockerHub, domain.RegistryTypeGenericV2:
		return true
	default:
		return false
	}
}
