// Package settings provides the postgres backend for storage.SettingsStorage -
// see internal/storage/storage.go.
package settings

import (
	"context"
	"database/sql"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/settings_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/whitelisted_images_queries"
)

type pgStorage struct {
	querier          *settings_queries.Queries
	whitelistQuerier *whitelisted_images_queries.Queries
}

func NewPg(db *sql.DB) storage.SettingsStorage {
	return &pgStorage{
		querier:          settings_queries.New(db),
		whitelistQuerier: whitelisted_images_queries.New(db),
	}
}

func (p *pgStorage) GetSettings(ctx context.Context) (domain.Settings, error) {
	row, err := p.querier.GetSettings(ctx)
	if err != nil {
		return domain.Settings{}, rerrors.Wrap(err, "error getting settings")
	}

	images, err := p.whitelistQuerier.ListWhitelistedImages(ctx)
	if err != nil {
		return domain.Settings{}, rerrors.Wrap(err, "error listing whitelisted images")
	}

	repositories := make([]string, 0, len(images))
	for _, image := range images {
		repositories = append(repositories, image.Repository)
	}

	settings := domain.Settings{
		IsSysboxEnabled:          row.IsSysboxEnabled,
		IsSysboxWhitelistIgnored: row.IsSysboxWhitelistIgnored,
		WhitelistedImages:        repositories,
	}

	return settings, nil
}

func (p *pgStorage) UpdateSettings(ctx context.Context, settings domain.Settings) (domain.Settings, error) {
	params := settings_queries.UpsertSettingsParams{
		IsSysboxEnabled:          settings.IsSysboxEnabled,
		IsSysboxWhitelistIgnored: settings.IsSysboxWhitelistIgnored,
	}

	row, err := p.querier.UpsertSettings(ctx, params)
	if err != nil {
		return domain.Settings{}, rerrors.Wrap(err, "error upserting settings")
	}

	updated := domain.Settings{
		IsSysboxEnabled:          row.IsSysboxEnabled,
		IsSysboxWhitelistIgnored: row.IsSysboxWhitelistIgnored,
	}

	return updated, nil
}
