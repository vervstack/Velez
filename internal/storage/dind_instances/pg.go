// Package dind_instances provides the postgres backend for
// storage.DindInstancesStorage - see internal/storage/storage.go.
package dind_instances

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/dind_instances_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

type pgStorage struct {
	querier *dind_instances_queries.Queries
}

func NewPg(db *sql.DB) storage.DindInstancesStorage {
	return &pgStorage{
		querier: dind_instances_queries.New(db),
	}
}

func (p *pgStorage) UpsertDindInstance(
	ctx context.Context, req domain.UpsertDindInstanceReq,
) (domain.DindInstance, error) {
	params := dind_instances_queries.UpsertDindInstanceParams{
		ServiceID:       req.ServiceId,
		IsSysboxEnabled: req.IsSysboxEnabled,
	}

	row, err := p.querier.UpsertDindInstance(ctx, params)
	if err != nil {
		return domain.DindInstance{}, rerrors.Wrap(wrapDindInstancesPgErr(err), "error upserting dind instance")
	}

	return dindInstanceFromRow(row), nil
}

func (p *pgStorage) GetDindInstanceByServiceId(
	ctx context.Context, serviceId int64,
) (domain.DindInstance, error) {
	row, err := p.querier.GetDindInstanceByServiceID(ctx, serviceId)
	if err != nil {
		return domain.DindInstance{}, rerrors.Wrap(
			wrapDindInstancesPgErr(err), "error getting dind instance by service id",
		)
	}

	return dindInstanceFromRow(row), nil
}

func (p *pgStorage) ListDindInstances(ctx context.Context) ([]domain.DindInstance, error) {
	rows, err := p.querier.ListDindInstances(ctx)
	if err != nil {
		return nil, rerrors.Wrap(wrapDindInstancesPgErr(err), "error listing dind instances")
	}

	out := make([]domain.DindInstance, 0, len(rows))
	for _, row := range rows {
		out = append(out, dindInstanceFromRow(row))
	}

	return out, nil
}

func (p *pgStorage) DeleteDindInstance(ctx context.Context, serviceId int64) error {
	err := p.querier.DeleteDindInstance(ctx, serviceId)
	if err != nil {
		return rerrors.Wrap(wrapDindInstancesPgErr(err), "error deleting dind instance")
	}

	return nil
}

func dindInstanceFromRow(row dind_instances_queries.VelezDindInstance) domain.DindInstance {
	return domain.DindInstance{
		ServiceId:       row.ServiceID,
		IsSysboxEnabled: row.IsSysboxEnabled,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

// wrapDindInstancesPgErr mirrors internal/storage/registry_instances
// .wrapRegistryInstancesPgErr - internal/storage/postgres.wrapPgErr can't be
// reused (import cycle).
func wrapDindInstancesPgErr(err error) error {
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
