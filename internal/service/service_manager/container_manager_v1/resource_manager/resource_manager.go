package resource_manager

import (
	"sync"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/client"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka/resources"

	"github.com/godverv/Velez/pkg/velez_api"
)

var ErrNotFound = errors.New("no resource with such name")

type SmerdConstructor func(resources matreshka.Resources, resourceName string) (*velez_api.CreateSmerd_Request, error)

type ResourceManager struct {
	docker client.CommonAPIClient

	m      sync.RWMutex
	images map[string]SmerdConstructor
}

func New(docker client.CommonAPIClient) *ResourceManager {
	r := &ResourceManager{
		docker: docker,
	}
	r.images = map[string]SmerdConstructor{
		resources.PostgresResourceName: Postgres,
	}

	return r
}

func (m *ResourceManager) GetByName(resourceName string) (SmerdConstructor, error) {
	m.m.RLock()
	launcher, ok := m.images[resourceName]
	m.m.RUnlock()

	if !ok {
		return nil, ErrNotFound
	}

	return launcher, nil
}
