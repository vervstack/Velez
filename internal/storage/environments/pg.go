package environments

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/environments_queries"
)

type pgStorage struct {
	querier *environments_queries.Queries
}

// NewPg builds a postgres-backed storage.EnvironmentsStorage on top of the
// sqlc-generated environments_queries package.
func NewPg(db *sql.DB) storage.EnvironmentsStorage {
	return &pgStorage{
		querier: environments_queries.New(db),
	}
}

func (p *pgStorage) ListEnvironments(ctx context.Context) ([]domain.Environment, error) {
	rows, err := p.querier.ListEnvironments(ctx)
	if err != nil {
		return nil, rerrors.Wrap(wrapEnvPgErr(err), "error listing environments")
	}

	out := make([]domain.Environment, 0, len(rows))
	for _, row := range rows {
		out = append(out, environmentFromRow(row))
	}

	return out, nil
}

func (p *pgStorage) GetEnvironmentByID(ctx context.Context, id int64) (domain.Environment, error) {
	row, err := p.querier.GetEnvironmentByID(ctx, id)
	if err != nil {
		return domain.Environment{}, rerrors.Wrap(wrapEnvPgErr(err), "error getting environment by id")
	}

	return environmentFromRow(row), nil
}

func (p *pgStorage) GetEnvironmentByName(ctx context.Context, name string) (domain.Environment, error) {
	row, err := p.querier.GetEnvironmentByName(ctx, name)
	if err != nil {
		return domain.Environment{}, rerrors.Wrap(wrapEnvPgErr(err), "error getting environment by name")
	}

	return environmentFromRow(row), nil
}

func (p *pgStorage) CreateEnvironment(
	ctx context.Context,
	req domain.CreateEnvironmentReq,
) (domain.Environment, error) {
	params := environments_queries.CreateEnvironmentParams{
		Name:   req.Name,
		Suffix: req.Suffix,
	}

	row, err := p.querier.CreateEnvironment(ctx, params)
	if err != nil {
		return domain.Environment{}, rerrors.Wrap(wrapEnvPgErr(err), "error creating environment")
	}

	return environmentFromRow(row), nil
}

func (p *pgStorage) UpdateEnvironment(
	ctx context.Context,
	req domain.UpdateEnvironmentReq,
) (domain.Environment, error) {
	current, err := p.querier.GetEnvironmentByID(ctx, req.ID)
	if err != nil {
		return domain.Environment{}, rerrors.Wrap(wrapEnvPgErr(err), "error getting environment before update")
	}

	params := environments_queries.UpdateEnvironmentParams{
		ID:     req.ID,
		Name:   current.Name,
		Suffix: current.Suffix,
	}

	if req.Name != nil {
		params.Name = *req.Name
	}

	if req.Suffix != nil {
		params.Suffix = *req.Suffix
	}

	row, err := p.querier.UpdateEnvironment(ctx, params)
	if err != nil {
		return domain.Environment{}, rerrors.Wrap(wrapEnvPgErr(err), "error updating environment")
	}

	return environmentFromRow(row), nil
}

func (p *pgStorage) DeleteEnvironment(ctx context.Context, id int64) error {
	err := p.querier.DeleteEnvironment(ctx, id)
	if err != nil {
		return rerrors.Wrap(wrapEnvPgErr(err), "error deleting environment")
	}

	return nil
}

func environmentFromRow(row environments_queries.VelezEnvironment) domain.Environment {
	return domain.Environment{
		ID:        row.ID,
		Name:      row.Name,
		Suffix:    row.Suffix,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

// wrapEnvPgErr mirrors internal/storage/postgres.wrapPgErr. It can't be reused
// directly - that helper is unexported and importing internal/storage/postgres
// from here would create an import cycle (postgres imports this package to
// build its Environments() storage).
func wrapEnvPgErr(err error) error {
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
