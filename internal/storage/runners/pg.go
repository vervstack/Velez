// Package runners provides a postgres backend for storage.RunnersStorage -
// see internal/storage/storage.go.
package runners

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/runners_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

type pgStorage struct {
	querier *runners_queries.Queries
}

// NewPg builds a postgres-backed storage.RunnersStorage on top of the
// sqlc-generated runners_queries package.
func NewPg(db *sql.DB) storage.RunnersStorage {
	return &pgStorage{
		querier: runners_queries.New(db),
	}
}

func (p *pgStorage) UpsertRunner(ctx context.Context, req domain.UpsertRunnerReq) (domain.Runner, error) {
	params := runners_queries.UpsertRunnerParams{
		ServiceID: req.ServiceID,
		Provider:  req.Provider,
		Scope:     req.Scope,
		Target:    req.Target,
		Labels:    req.Labels,
		SecretRef: req.SecretRef,
	}

	row, err := p.querier.UpsertRunner(ctx, params)
	if err != nil {
		return domain.Runner{}, rerrors.Wrap(wrapRunnersPgErr(err), "error upserting runner")
	}

	return runnerFromRow(row), nil
}

func (p *pgStorage) GetRunnerByServiceID(ctx context.Context, serviceID int64) (domain.Runner, error) {
	row, err := p.querier.GetRunnerByServiceID(ctx, serviceID)
	if err != nil {
		return domain.Runner{}, rerrors.Wrap(wrapRunnersPgErr(err), "error getting runner by service id")
	}

	return runnerFromRow(row), nil
}

func (p *pgStorage) ListRunners(ctx context.Context) ([]domain.Runner, error) {
	rows, err := p.querier.ListRunners(ctx)
	if err != nil {
		return nil, rerrors.Wrap(wrapRunnersPgErr(err), "error listing runners")
	}

	out := make([]domain.Runner, 0, len(rows))
	for _, row := range rows {
		out = append(out, runnerFromRow(row))
	}

	return out, nil
}

func (p *pgStorage) DeleteRunner(ctx context.Context, serviceID int64) error {
	err := p.querier.DeleteRunner(ctx, serviceID)
	if err != nil {
		return rerrors.Wrap(wrapRunnersPgErr(err), "error deleting runner")
	}

	return nil
}

func runnerFromRow(row runners_queries.VelezRunner) domain.Runner {
	return domain.Runner{
		ServiceID: row.ServiceID,
		Provider:  row.Provider,
		Scope:     row.Scope,
		Target:    row.Target,
		Labels:    row.Labels,
		SecretRef: row.SecretRef,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

// wrapRunnersPgErr mirrors internal/storage/pg_instances.wrapPgInstancesPgErr
// / internal/storage/postgres.wrapPgErr. It can't be reused directly - those
// helpers are unexported and importing internal/storage/postgres from here
// would create an import cycle (postgres imports this package to build its
// Runners() storage).
func wrapRunnersPgErr(err error) error {
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
