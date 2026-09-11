// Package pg_instances provides postgres and in-memory backends for
// storage.PgInstancesStorage - see internal/storage/storage.go.
package pg_instances

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/pg_instances_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

type pgStorage struct {
	querier *pg_instances_queries.Queries
}

// NewPg builds a postgres-backed storage.PgInstancesStorage on top of the
// sqlc-generated pg_instances_queries package.
func NewPg(db *sql.DB) storage.PgInstancesStorage {
	return &pgStorage{
		querier: pg_instances_queries.New(db),
	}
}

func (p *pgStorage) UpsertPgInstance(ctx context.Context, req domain.UpsertPgInstanceReq) (domain.PgInstance, error) {
	params := pg_instances_queries.UpsertPgInstanceParams{
		ServiceID: req.ServiceID,
		DbName:    req.DbName,
		Username:  req.Username,
		SecretRef: req.SecretRef,
		Port:      req.Port,
	}

	row, err := p.querier.UpsertPgInstance(ctx, params)
	if err != nil {
		return domain.PgInstance{}, rerrors.Wrap(wrapPgInstancesPgErr(err), "error upserting pg instance")
	}

	return pgInstanceFromRow(row), nil
}

func (p *pgStorage) GetPgInstanceByServiceID(ctx context.Context, serviceID int64) (domain.PgInstance, error) {
	row, err := p.querier.GetPgInstanceByServiceID(ctx, serviceID)
	if err != nil {
		return domain.PgInstance{}, rerrors.Wrap(wrapPgInstancesPgErr(err), "error getting pg instance by service id")
	}

	return pgInstanceFromRow(row), nil
}

func (p *pgStorage) ListPgInstances(ctx context.Context) ([]domain.PgInstance, error) {
	rows, err := p.querier.ListPgInstances(ctx)
	if err != nil {
		return nil, rerrors.Wrap(wrapPgInstancesPgErr(err), "error listing pg instances")
	}

	out := make([]domain.PgInstance, 0, len(rows))
	for _, row := range rows {
		out = append(out, pgInstanceFromRow(row))
	}

	return out, nil
}

func (p *pgStorage) DeletePgInstance(ctx context.Context, serviceID int64) error {
	err := p.querier.DeletePgInstance(ctx, serviceID)
	if err != nil {
		return rerrors.Wrap(wrapPgInstancesPgErr(err), "error deleting pg instance")
	}

	return nil
}

func pgInstanceFromRow(row pg_instances_queries.VelezPgInstance) domain.PgInstance {
	return domain.PgInstance{
		ServiceID: row.ServiceID,
		DbName:    row.DbName,
		Username:  row.Username,
		SecretRef: row.SecretRef,
		Port:      row.Port,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

// wrapPgInstancesPgErr mirrors internal/storage/postgres.wrapPgErr /
// internal/storage/registries.wrapRegistryPgErr. It can't be reused directly -
// those helpers are unexported and importing internal/storage/postgres from
// here would create an import cycle (postgres imports this package to build
// its PgInstances() storage).
func wrapPgInstancesPgErr(err error) error {
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
