// Package registry_instances provides the postgres backend for
// storage.RegistryInstancesStorage - see internal/storage/storage.go.
package registry_instances

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/registry_instances_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

type pgStorage struct {
	querier *registry_instances_queries.Queries
}

// NewPg builds a postgres-backed storage.RegistryInstancesStorage on top of
// the sqlc-generated registry_instances_queries package.
func NewPg(db *sql.DB) storage.RegistryInstancesStorage {
	return &pgStorage{
		querier: registry_instances_queries.New(db),
	}
}

func (p *pgStorage) UpsertRegistryInstance(
	ctx context.Context, req domain.UpsertRegistryInstanceReq,
) (domain.RegistryInstance, error) {
	params := registry_instances_queries.UpsertRegistryInstanceParams{
		ServiceID: req.ServiceId,
		Port:      req.Port,
		UiPort:    req.UiPort,
		Username:  req.Username,
		SecretRef: req.SecretRef,
	}

	row, err := p.querier.UpsertRegistryInstance(ctx, params)
	if err != nil {
		return domain.RegistryInstance{}, rerrors.Wrap(wrapRegistryInstancesPgErr(err), "error upserting registry instance")
	}

	return registryInstanceFromRow(row), nil
}

func (p *pgStorage) GetRegistryInstanceByServiceID(
	ctx context.Context, serviceID int64,
) (domain.RegistryInstance, error) {
	row, err := p.querier.GetRegistryInstanceByServiceID(ctx, serviceID)
	if err != nil {
		return domain.RegistryInstance{}, rerrors.Wrap(
			wrapRegistryInstancesPgErr(err), "error getting registry instance by service id",
		)
	}

	return registryInstanceFromRow(row), nil
}

func (p *pgStorage) ListRegistryInstances(ctx context.Context) ([]domain.RegistryInstance, error) {
	rows, err := p.querier.ListRegistryInstances(ctx)
	if err != nil {
		return nil, rerrors.Wrap(wrapRegistryInstancesPgErr(err), "error listing registry instances")
	}

	out := make([]domain.RegistryInstance, 0, len(rows))
	for _, row := range rows {
		out = append(out, registryInstanceFromRow(row))
	}

	return out, nil
}

func (p *pgStorage) DeleteRegistryInstance(ctx context.Context, serviceID int64) error {
	err := p.querier.DeleteRegistryInstance(ctx, serviceID)
	if err != nil {
		return rerrors.Wrap(wrapRegistryInstancesPgErr(err), "error deleting registry instance")
	}

	return nil
}

func registryInstanceFromRow(row registry_instances_queries.VelezRegistryInstance) domain.RegistryInstance {
	return domain.RegistryInstance{
		ServiceId: row.ServiceID,
		Port:      row.Port,
		UiPort:    row.UiPort,
		Username:  row.Username,
		SecretRef: row.SecretRef,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

// wrapRegistryInstancesPgErr mirrors internal/storage/pg_instances
// .wrapPgInstancesPgErr - can't be reused directly, see that function's own
// doc comment for why.
func wrapRegistryInstancesPgErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return rerrors.Wrap(user_errors.ErrStorageNotFound)
	}

	var pgErr *pq.Error

	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" { // unique_violation
			return errors.Join(user_errors.ErrStorageAlreadyExists, err)
		}
	}

	return err
}
