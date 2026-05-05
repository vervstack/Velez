package postgres

import (
	"context"
	"database/sql"

	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/plugins_queries"
)

type pluginsStorage struct {
	querier plugins_queries.Querier
}

func newPluginsStorage(db *sql.DB) *pluginsStorage {
	return &pluginsStorage{
		querier: plugins_queries.New(db),
	}
}

func (p *pluginsStorage) ListPlugins(ctx context.Context) ([]plugins_queries.ListPluginsRow, error) {
	return p.querier.ListPlugins(ctx)
}

func (p *pluginsStorage) UpsertPlugin(ctx context.Context, arg plugins_queries.UpsertPluginParams) error {
	return p.querier.UpsertPlugin(ctx, arg)
}

var _ storage.PluginsStorage = (*pluginsStorage)(nil)
