package local_storage

import (
	"context"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/plugins_queries"
)

type plugins struct{}

func newPluginsStorage() *plugins {
	return &plugins{}
}

func (p *plugins) ListPlugins(_ context.Context) ([]domain.PluginBaseInfo, error) {
	return nil, nil
}

func (p *plugins) UpsertPlugin(_ context.Context, _ plugins_queries.UpsertPluginParams) error {
	return nil
}
