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
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
)

// PostgresServiceImpl implements service.PostgresService.
type PostgresServiceImpl struct {
	dataStorage storage.Storage

	vervServices service.VervServicesService
	secrets      secrets.Store
	jobsEngine   jobs.Engine
	runtimes     container_runtime.RuntimeResolver
	docker       node_clients.Docker
}

func New(
	dataStorage storage.Storage,
	vervServices service.VervServicesService,
	secretsStore secrets.Store,
	jobsEngine jobs.Engine,
	runtimes container_runtime.RuntimeResolver,
	docker node_clients.Docker,
) *PostgresServiceImpl {
	return &PostgresServiceImpl{
		dataStorage:  dataStorage,
		vervServices: vervServices,
		secrets:      secretsStore,
		jobsEngine:   jobsEngine,
		runtimes:     runtimes,
		docker:       docker,
	}
}
