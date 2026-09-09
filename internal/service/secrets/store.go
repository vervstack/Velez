// Package secrets implements the seam
// docs/features/pgaas_and_registry_plugin.md section 1 calls "the Svarog
// fallback": a deliberately poor secret store that a future Svarog (a Vault
// alternative) service will replace. Nothing outside this package may touch
// storage.SecretsStorage / velez.secrets directly.
package secrets

import (
	"context"
	"errors"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// Store is the seam that lets a future svarogStore implementation replace
// dbStore without changing any caller.
type Store interface {
	Put(ctx context.Context, ref domain.SecretRef, value string) error
	Get(ctx context.Context, ref domain.SecretRef) (string, error)
	Delete(ctx context.Context, ref domain.SecretRef) error
	// ListRefs returns keys only - never values.
	ListRefs(ctx context.Context, scope, owner string) ([]domain.SecretRef, error)
}

// dbStore is the only Store implementation today, over
// storage.SecretsStorage (postgres + static backends - see
// internal/storage/secrets). A future svarogStore will implement Store the
// same way; internal/app/custom.go picks between them.
type dbStore struct {
	storage storage.Storage
}

// New builds a Store backed by velez.secrets.
func New(stg storage.Storage) Store {
	return &dbStore{storage: stg}
}

func (d *dbStore) Put(ctx context.Context, ref domain.SecretRef, value string) error {
	err := d.storage.Secrets().PutSecret(ctx, ref, value)
	if err != nil {
		return rerrors.Wrap(err, "error putting secret")
	}

	return nil
}

func (d *dbStore) Get(ctx context.Context, ref domain.SecretRef) (string, error) {
	value, err := d.storage.Secrets().GetSecret(ctx, ref)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", rerrors.Wrap(user_errors.ErrSecretNotFound)
		}

		return "", rerrors.Wrap(err, "error getting secret")
	}

	return value, nil
}

func (d *dbStore) Delete(ctx context.Context, ref domain.SecretRef) error {
	err := d.storage.Secrets().DeleteSecret(ctx, ref)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return rerrors.Wrap(user_errors.ErrSecretNotFound)
		}

		return rerrors.Wrap(err, "error deleting secret")
	}

	return nil
}

func (d *dbStore) ListRefs(ctx context.Context, scope, owner string) ([]domain.SecretRef, error) {
	refs, err := d.storage.Secrets().ListSecretRefs(ctx, scope, owner)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing secret refs")
	}

	return refs, nil
}
