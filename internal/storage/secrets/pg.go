// Package secrets provides postgres and in-memory backends for
// storage.SecretsStorage - see internal/storage/storage.go.
package secrets

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/secrets_queries"
)

type pgStorage struct {
	querier *secrets_queries.Queries
}

// NewPg builds a postgres-backed storage.SecretsStorage on top of the
// sqlc-generated secrets_queries package.
func NewPg(db *sql.DB) storage.SecretsStorage {
	return &pgStorage{
		querier: secrets_queries.New(db),
	}
}

func (p *pgStorage) PutSecret(ctx context.Context, ref domain.SecretRef, value string) error {
	params := secrets_queries.PutSecretParams{
		Scope: ref.Scope,
		Owner: ref.Owner,
		Key:   ref.Key,
		Value: value,
	}

	err := p.querier.PutSecret(ctx, params)
	if err != nil {
		return rerrors.Wrap(wrapSecretsPgErr(err), "error putting secret")
	}

	return nil
}

func (p *pgStorage) GetSecret(ctx context.Context, ref domain.SecretRef) (string, error) {
	params := secrets_queries.GetSecretValueParams{
		Scope: ref.Scope,
		Owner: ref.Owner,
		Key:   ref.Key,
	}

	value, err := p.querier.GetSecretValue(ctx, params)
	if err != nil {
		return "", rerrors.Wrap(wrapSecretsPgErr(err), "error getting secret")
	}

	return value, nil
}

func (p *pgStorage) DeleteSecret(ctx context.Context, ref domain.SecretRef) error {
	params := secrets_queries.DeleteSecretParams{
		Scope: ref.Scope,
		Owner: ref.Owner,
		Key:   ref.Key,
	}

	err := p.querier.DeleteSecret(ctx, params)
	if err != nil {
		return rerrors.Wrap(wrapSecretsPgErr(err), "error deleting secret")
	}

	return nil
}

func (p *pgStorage) ListSecretRefs(ctx context.Context, scope, owner string) ([]domain.SecretRef, error) {
	params := secrets_queries.ListSecretRefsParams{
		Scope: scope,
		Owner: owner,
	}

	rows, err := p.querier.ListSecretRefs(ctx, params)
	if err != nil {
		return nil, rerrors.Wrap(wrapSecretsPgErr(err), "error listing secret refs")
	}

	out := make([]domain.SecretRef, 0, len(rows))
	for _, row := range rows {
		ref := domain.SecretRef{
			Scope: row.Scope,
			Owner: row.Owner,
			Key:   row.Key,
		}

		out = append(out, ref)
	}

	return out, nil
}

// wrapSecretsPgErr mirrors internal/storage/postgres.wrapPgErr /
// internal/storage/registries.wrapRegistryPgErr. It can't be reused directly -
// those helpers are unexported and importing internal/storage/postgres from
// here would create an import cycle (postgres imports this package to build
// its Secrets() storage).
func wrapSecretsPgErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return rerrors.Wrap(storage.ErrNotFound)
	}

	var pgErr *pq.Error

	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" { // unique_violation
			return errors.Join(storage.ErrAlreadyExists, err)
		}
	}

	return err
}
