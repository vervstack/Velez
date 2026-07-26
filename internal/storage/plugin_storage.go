package storage

import (
	"sync/atomic"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
)

type Container struct {
	impl atomic.Pointer[Storage]
}

func NewStorageContainer(initial Storage) *Container {
	c := &Container{}
	c.Set(initial)

	return c
}

func (c *Container) Set(impl Storage) {
	c.impl.Store(&impl)
}

func (c *Container) Nodes() NodesStorage {
	return (*c.impl.Load()).Nodes()
}

func (c *Container) Services() ServicesStorage {
	return (*c.impl.Load()).Services()
}

func (c *Container) Deployments() DeploymentsStorage {
	return (*c.impl.Load()).Deployments()
}

func (c *Container) Plugins() PluginsStorage {
	return (*c.impl.Load()).Plugins()
}

func (c *Container) TxManager() *sqldb.TxManager {
	return (*c.impl.Load()).TxManager()
}

func (c *Container) ServiceDependencies() ServiceDependenciesStorage {
	return (*c.impl.Load()).ServiceDependencies()
}

func (c *Container) ServiceResources() ServiceResourcesStorage {
	return (*c.impl.Load()).ServiceResources()
}

func (c *Container) Tasks() TasksStorage {
	return (*c.impl.Load()).Tasks()
}

func (c *Container) Jobs() JobsStorage {
	return (*c.impl.Load()).Jobs()
}
