package registries

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/registries_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

type pgStorage struct {
	querier *registries_queries.Queries
}

// NewPg builds a postgres-backed storage.RegistriesStorage on top of the
// sqlc-generated registries_queries package.
func NewPg(db *sql.DB) storage.RegistriesStorage {
	return &pgStorage{
		querier: registries_queries.New(db),
	}
}

func (p *pgStorage) ListRegistries(ctx context.Context) ([]domain.Registry, error) {
	rows, err := p.querier.ListRegistries(ctx)
	if err != nil {
		return nil, rerrors.Wrap(wrapRegistryPgErr(err), "error listing registries")
	}

	out := make([]domain.Registry, 0, len(rows))
	for _, row := range rows {
		out = append(out, registryFromRow(row))
	}

	return out, nil
}

func (p *pgStorage) GetRegistryByID(ctx context.Context, id int64) (domain.Registry, error) {
	row, err := p.querier.GetRegistryByID(ctx, id)
	if err != nil {
		return domain.Registry{}, rerrors.Wrap(wrapRegistryPgErr(err), "error getting registry by id")
	}

	return registryFromRow(row), nil
}

func (p *pgStorage) CreateRegistry(ctx context.Context, req domain.CreateRegistryReq) (domain.Registry, error) {
	params := registries_queries.CreateRegistryParams{
		Name:      req.Name,
		Type:      string(req.Type),
		Url:       req.Url,
		Username:  req.Username,
		Secret:    req.Secret,
		IsDefault: req.IsDefault,
	}

	row, err := p.querier.CreateRegistry(ctx, params)
	if err != nil {
		return domain.Registry{}, rerrors.Wrap(wrapRegistryPgErr(err), "error creating registry")
	}

	return registryFromRow(row), nil
}

func (p *pgStorage) UpdateRegistry(ctx context.Context, req domain.UpdateRegistryReq) (domain.Registry, error) {
	current, err := p.querier.GetRegistryByID(ctx, req.Id)
	if err != nil {
		return domain.Registry{}, rerrors.Wrap(wrapRegistryPgErr(err), "error getting registry before update")
	}

	params := registries_queries.UpdateRegistryParams{
		ID:        req.Id,
		Name:      current.Name,
		Type:      current.Type,
		Url:       current.Url,
		Username:  current.Username,
		Secret:    current.Secret,
		IsDefault: current.IsDefault,
	}

	if req.Name != nil {
		params.Name = *req.Name
	}

	if req.Type != nil {
		params.Type = string(*req.Type)
	}

	if req.Url != nil {
		params.Url = *req.Url
	}

	if req.Username != nil {
		params.Username = *req.Username
	}

	if req.Secret != nil {
		params.Secret = *req.Secret
	}

	if req.IsDefault != nil {
		params.IsDefault = *req.IsDefault
	}

	row, err := p.querier.UpdateRegistry(ctx, params)
	if err != nil {
		return domain.Registry{}, rerrors.Wrap(wrapRegistryPgErr(err), "error updating registry")
	}

	return registryFromRow(row), nil
}

func (p *pgStorage) DeleteRegistry(ctx context.Context, id int64) error {
	err := p.querier.DeleteRegistry(ctx, id)
	if err != nil {
		return rerrors.Wrap(wrapRegistryPgErr(err), "error deleting registry")
	}

	return nil
}

func (p *pgStorage) ClearDefaultRegistry(ctx context.Context) error {
	err := p.querier.ClearDefaultRegistry(ctx)
	if err != nil {
		return rerrors.Wrap(wrapRegistryPgErr(err), "error clearing default registry")
	}

	return nil
}

func (p *pgStorage) WithTx(tx *sql.Tx) storage.RegistriesStorage {
	return &pgStorage{
		querier: p.querier.WithTx(tx),
	}
}

func registryFromRow(row registries_queries.VelezRegistry) domain.Registry {
	return domain.Registry{
		Id:        row.ID,
		Name:      row.Name,
		Type:      domain.RegistryType(row.Type),
		Url:       row.Url,
		Username:  row.Username,
		Secret:    row.Secret,
		IsDefault: row.IsDefault,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

// wrapRegistryPgErr mirrors internal/storage/postgres.wrapPgErr /
// internal/storage/environments.wrapEnvPgErr. It can't be reused directly -
// those helpers are unexported and importing internal/storage/postgres from
// here would create an import cycle (postgres imports this package to build
// its Registries() storage).
func wrapRegistryPgErr(err error) error {
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
