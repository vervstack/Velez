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
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
)

// PgaasService implements service.PostgresService.
type PgaasService struct {
	dataStorage storage.Storage

	vervServices service.VervServicesService
	secrets      secrets.Store
	jobsEngine   jobs.Engine
}

func New(
	dataStorage storage.Storage,
	vervServices service.VervServicesService,
	secretsStore secrets.Store,
	jobsEngine jobs.Engine,
) *PgaasService {
	return &PgaasService{
		dataStorage:  dataStorage,
		vervServices: vervServices,
		secrets:      secretsStore,
		jobsEngine:   jobsEngine,
	}
}
