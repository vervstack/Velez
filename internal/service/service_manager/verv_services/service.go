package verv_services

import (
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage"
)

type VervService struct {
	servicesStorage    storage.ServicesStorage
	deploymentsStorage storage.DeploymentsStorage

	txManager *sqldb.TxManager
}

func New(dataStorage storage.Storage) *VervService {
	return &VervService{
		servicesStorage:    dataStorage.Services(),
		deploymentsStorage: dataStorage.Deployments(),

		txManager: dataStorage.TxManager(),
	}
}
