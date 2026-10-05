// Package s3aas implements service.S3Service - Garage-backed S3 instances.
// Velez keeps no per-instance table: the instance containers' labels and port
// bindings are the source of truth, and buckets and keys live in Garage.
package s3aas

import (
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
)

type Service struct {
	dataStorage    storage.Storage
	vervServices   service.VervServicesService
	secretsStore   secrets.Store
	configResolver service.ServiceConfigResolver
	jobsEngine     jobs.Engine
	runtimes       container_runtime.RuntimeResolver
	docker         node_clients.Docker
}

func New(
	dataStorage storage.Storage,
	vervServices service.VervServicesService,
	secretsStore secrets.Store,
	configResolver service.ServiceConfigResolver,
	jobsEngine jobs.Engine,
	runtimes container_runtime.RuntimeResolver,
	docker node_clients.Docker,
) *Service {
	return &Service{
		dataStorage:    dataStorage,
		vervServices:   vervServices,
		secretsStore:   secretsStore,
		configResolver: configResolver,
		jobsEngine:     jobsEngine,
		runtimes:       runtimes,
		docker:         docker,
	}
}
