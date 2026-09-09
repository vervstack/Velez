// Package pgaas implements service.PostgresService - Postgres-as-a-Service.
// A PG instance is a normal Velez service: it gets velez.services,
// deployment_specifications and deployments rows through the ordinary
// VervServicesService.CreateNewDeploy path, and the existing deploy watcher
// creates the container. Only the pg-specific facts (db name, username,
// secret ref, port) get their own linked row, velez.pg_instances - status,
// environment and image are always read back through VervServicesService,
// never duplicated. See docs/features/pgaas_and_registry_plugin.md section 3.
package pgaas

import (
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	// pgResourceType is the velez.service_resources.resource_type stamped
	// onto the (owner, instance) binding CreatePgInstance records when an
	// owner service is given.
	pgResourceType = "postgres"

	// pgDefaultPort is the port a PG instance's container listens on
	// internally - fixed by the builtin postgres descriptor
	// (builtin/postgres/deployment.yaml), independent of whatever host port
	// ExposeToPort publishes it on.
	pgDefaultPort = 5432
)

// PgaasService implements service.PostgresService.
type PgaasService struct {
	dataStorage storage.Storage

	vervServices service.VervServicesService
	secrets      secrets.Store
}

// New builds a PgaasService. dataStorage, vervServices and secretsStore are
// interfaces, never concrete types - dataStorage is resolved per call
// (mirroring verv_services.VervService's boxes()/environments()) so a
// runtime storage backend swap is picked up immediately.
func New(
	dataStorage storage.Storage, vervServices service.VervServicesService, secretsStore secrets.Store,
) *PgaasService {
	return &PgaasService{
		dataStorage:  dataStorage,
		vervServices: vervServices,
		secrets:      secretsStore,
	}
}

// boxes mirrors verv_services.VervService.boxes() - resolved per call so a
// runtime storage swap is picked up immediately.
func (s *PgaasService) boxes() storage.ResourceBoxesStorage {
	return s.dataStorage.ResourceBoxes()
}
