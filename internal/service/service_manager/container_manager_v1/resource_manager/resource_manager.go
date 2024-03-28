package resource_manager

import (
	"encoding/gob"
	"sync"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/godverv/matreshka/resources"

	"github.com/godverv/Velez/pkg/velez_api"
)

var ErrNotFound = errors.New("no resource with such name")

type ResourceManager struct {
	docker client.CommonAPIClient

	m      sync.RWMutex
	images map[string]velez_api.CreateSmerd_Request
}

func New(docker client.CommonAPIClient) *ResourceManager {
	r := &ResourceManager{
		docker: docker,
	}
	r.images = map[string]velez_api.CreateSmerd_Request{
		resources.PostgresResourceName: {
			ImageName: "postgres:13.6",
			Hardware:  &velez_api.Container_Hardware{},
			Settings:  &velez_api.Container_Settings{},
		},
	}

	return r
}

func (m *ResourceManager) GetByName(in string) (*velez_api.CreateSmerd_Request, error) {
	m.m.RLock()
	launcher, ok := m.images[in]
	m.m.RUnlock()

	pipe := ioutils.NewBytesPipe()

	enc := gob.NewEncoder(pipe)
	err := enc.Encode(&launcher)
	if err != nil {
		return nil, errors.Wrap(err, "error encoding request")
	}

	var copied velez_api.CreateSmerd_Request
	err = gob.NewDecoder(pipe).Decode(&copied)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding request")
	}

	if ok {
		return &copied, nil
	}

	return nil, ErrNotFound
}
