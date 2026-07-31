package verv_services

import (
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/storage"
)

type VervService struct {
	dataStorage storage.Storage

	containerService    service.ContainerService
	docker              node_clients.Docker
	environmentsStorage storage.EnvironmentsStorageContainer
}

func New(
	dataStorage storage.Storage,
	containerService service.ContainerService,
	docker node_clients.Docker,
	environmentsStorage storage.EnvironmentsStorageContainer,
) *VervService {
	return &VervService{
		dataStorage: dataStorage,

		containerService:    containerService,
		docker:              docker,
		environmentsStorage: environmentsStorage,
	}
}
