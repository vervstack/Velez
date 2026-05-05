package storage

import (
	"context"
	"sync/atomic"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/plugins_queries"
)

type PluginsStorageContainer struct {
	impl atomic.Pointer[PluginsStorage]
}

func NewPluginsStorageContainer(initial PluginsStorage) *PluginsStorageContainer {
	c := &PluginsStorageContainer{}
	c.Set(initial)
	return c
}

func (c *PluginsStorageContainer) Set(impl PluginsStorage) {
	c.impl.Store(&impl)
}

func (c *PluginsStorageContainer) ListPlugins(ctx context.Context) ([]domain.PluginBaseInfo, error) {
	l := c.impl.Load()
	return (*l).ListPlugins(ctx)
}

func (c *PluginsStorageContainer) UpsertPlugin(ctx context.Context, arg plugins_queries.UpsertPluginParams) error {
	l := c.impl.Load()
	return (*l).UpsertPlugin(ctx, arg)
}
