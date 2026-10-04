// Package dinds implements service.DindService - Docker-in-Docker daemons. A
// DinD is a normal Velez service; only the dind-specific facts get their own
// linked row, velez.dind_instances.
package dinds

import (
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/storage"
)

type Service struct {
	dataStorage  storage.Storage
	vervServices service.VervServicesService
	jobsEngine   jobs.Engine
	docker       node_clients.Docker
}

func New(
	dataStorage storage.Storage, vervServices service.VervServicesService,
	jobsEngine jobs.Engine, docker node_clients.Docker,
) *Service {
	return &Service{
		dataStorage:  dataStorage,
		vervServices: vervServices,
		jobsEngine:   jobsEngine,
		docker:       docker,
	}
}
