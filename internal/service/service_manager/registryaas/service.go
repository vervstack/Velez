// Package registryaas implements service.ContainerRegistryService -
// Container-Registry-as-a-Service. A registry instance is a normal Velez
// service - two of them, in fact: the docker registry:2 container itself and
// a docker-registry-ui sidecar - deployed through the ordinary
// VervServicesService.CreateNewDeploy path from the builtin registry/
// registry_ui vervonomicon descriptors. Unlike PostgresService, creation is a
// multi-step operation (credentials, port resolution, an htpasswd loader
// container, two deploys, two storage rows) so it runs through the jobs
// engine (internal/jobs.CreateRegistryInstanceAction) rather than being done
// inline - CreateRegistryInstance enqueues the task and watches it to a
// terminal status, so the RPC still returns the fully-built instance
// synchronously. See docs/features/pgaas_and_registry_plugin.md section 4.
package registryaas

import (
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
)

// RegistryaasService implements service.ContainerRegistryService.
type RegistryaasService struct {
	dataStorage storage.Storage

	vervServices service.VervServicesService
	secrets      secrets.Store
	jobsEngine   jobs.Engine
}

// New builds a RegistryaasService. dataStorage, vervServices, secretsStore
// and jobsEngine are interfaces, never concrete types - mirrors pgaas.New.
func New(
	dataStorage storage.Storage,
	vervServices service.VervServicesService,
	secretsStore secrets.Store,
	jobsEngine jobs.Engine,
) *RegistryaasService {
	return &RegistryaasService{
		dataStorage:  dataStorage,
		vervServices: vervServices,
		secrets:      secretsStore,
		jobsEngine:   jobsEngine,
	}
}
