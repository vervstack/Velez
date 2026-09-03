package verv_services

import (
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/storage"
)

type VervService struct {
	dataStorage storage.Storage

	containerService service.ContainerService
	docker           node_clients.Docker

	// runtimes resolves the ContainerRuntime serving a given environment -
	// see docs/container_runtimes. Stop/Restart/Remove/GetServiceMetrics
	// route through it (instead of calling docker directly) so they resolve
	// against the smerd's actual environment rather than always PROD - see
	// docs/container_runtimes/roadmap.md's Stage 2.
	runtimes container_runtime.RuntimeResolver

	// vervSource reads .verv/ descriptor files out of a service's image, for
	// GetVervonomicon and CreateDeployFromVervonomicon.
	vervSource service.VervonomiconSource

	// configService reads a service's live matreshka config, for resource
	// reconciliation (GetVervonomicon and CreateDeployFromVervonomicon) - see
	// docs/features/vervonomicon.md's "Resource reconciliation".
	configService service.ConfigurationService
}

func New(
	dataStorage storage.Storage,
	containerService service.ContainerService,
	docker node_clients.Docker,
	runtimes container_runtime.RuntimeResolver,
	vervSource service.VervonomiconSource,
	configService service.ConfigurationService,
) *VervService {
	return &VervService{
		dataStorage: dataStorage,

		containerService: containerService,
		docker:           docker,
		runtimes:         runtimes,
		vervSource:       vervSource,
		configService:    configService,
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

// registries mirrors environments() - resolved per call so a runtime storage
// swap is picked up immediately.
func (v *VervService) registries() storage.RegistriesStorage {
	return v.dataStorage.Registries()
}

// boxes mirrors environments() - resolved per call so a runtime storage swap
// is picked up immediately.
func (v *VervService) boxes() storage.ResourceBoxesStorage {
	return v.dataStorage.ResourceBoxes()
}
