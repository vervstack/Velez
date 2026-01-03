package verv_services

import (
	"go.vervstack.ru/Velez/internal/storage"
)

type VervService struct {
	servicesStorage    storage.ServicesStorage
	deploymentsStorage storage.DeploymentsStorage
}

func New(dataStorage storage.Storage) *VervService {
	return &VervService{
		servicesStorage: dataStorage.Services(),
	}
}
