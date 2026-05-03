package verv_services

import (
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/storage"
)

type VervService struct {
	servicesStorage    storage.ServicesStorage
	deploymentsStorage storage.DeploymentsStorage

	txManager *sqldb.TxManager

	containerService service.ContainerService
	docker           node_clients.Docker
}

func New(dataStorage storage.Storage, containerService service.ContainerService, docker node_clients.Docker) *VervService {
	return &VervService{
		servicesStorage:    dataStorage.Services(),
		deploymentsStorage: dataStorage.Deployments(),

		txManager: dataStorage.TxManager(),

		containerService: containerService,
		docker:           docker,
	}
}
