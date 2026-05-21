package environments

import (
	"context"
	"sync/atomic"

	"go.vervstack.ru/Velez/internal/storage"
)

type envContainer struct {
	state atomic.Pointer[storage.EnvironmentsStorage]
}

func NewContainer(init storage.EnvironmentsStorage) storage.EnvironmentsStorageContainer {
	c := &envContainer{}
	c.Set(init)
	return c
}

func (c *envContainer) Set(s storage.EnvironmentsStorage) {
	c.state.Store(&s)
}

func (c *envContainer) ListEnvironments(ctx context.Context) ([]string, error) {
	return (*c.state.Load()).ListEnvironments(ctx)
}
