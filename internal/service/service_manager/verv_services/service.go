package verv_services

import (
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/storage"
)

type VervService struct {
	dataStorage storage.Storage

	containerService service.ContainerService
	docker           node_clients.Docker
}

func New(
	dataStorage storage.Storage,
	containerService service.ContainerService,
	docker node_clients.Docker,
) *VervService {
	return &VervService{
		dataStorage: dataStorage,

		containerService: containerService,
		docker:           docker,
	}
}

// environments returns the environments storage of whatever storage backend is
// currently live. It's resolved per call (rather than cached in a field) so a
// runtime swap from local_storage to postgres - see
// cluster_clients.ClusterStateManagerContainer.Set - is picked up immediately.
// This replaces the old, separate storage.EnvironmentsStorageContainer, which
// duplicated the swap mechanism the state manager container already provides.
func (v *VervService) environments() storage.EnvironmentsStorage {
	return v.dataStorage.Environments()
}
