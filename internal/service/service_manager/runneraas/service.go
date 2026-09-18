// Package runneraas implements service.RunnersService - Runners-as-a-Service.
// A runner is a normal Velez service: it gets velez.services,
// deployment_specifications and deployments rows through the ordinary
// VervServicesService.CreateNewDeploy path, and the existing deploy watcher
// creates the container. Only the runner-specific facts (provider, scope,
// target, labels, secret ref) get their own linked row, velez.runners -
// status and environment are always read back through VervServicesService,
// never duplicated. GitHub Actions is the first RunnerProvider; everything in
// this package outside of Provider selection is git-agnostic.
package runneraas

import (
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
)

// RunneraasService implements service.RunnersService - see runners_api.proto's
// RunnersAPI doc comment. A runner is a normal Velez service, deployed
// through the ordinary VervServicesService.CreateNewDeploy path; only
// provider/scope/target/labels/secret_ref get their own linked row,
// velez.runners - status and environment are always read back through
// VervServicesService, never duplicated. Provider-specific behavior (minting
// a registration token, its env-var vocabulary) is delegated to a
// providers.Provider, selected by RunnerProvider - everything else here is
// git-agnostic. Creating a runner is a multi-step operation (mint token,
// deploy, wait for the container, register the row), so like registryaas it
// runs through the jobs engine (internal/jobs.CreateRunnerAction) rather than
// being done inline - see create.go.
type RunneraasService struct {
	dataStorage  storage.Storage
	vervServices service.VervServicesService
	secrets      secrets.Store
	jobsEngine   jobs.Engine
}

// New builds a RunneraasService. dataStorage, vervServices, secretsStore and
// jobsEngine are interfaces, never concrete types - dataStorage is resolved
// per call (mirroring pgaas.PgaasService) so a runtime storage backend swap
// is picked up immediately.
func New(
	dataStorage storage.Storage, vervServices service.VervServicesService, secretsStore secrets.Store,
	jobsEngine jobs.Engine,
) *RunneraasService {
	return &RunneraasService{
		dataStorage:  dataStorage,
		vervServices: vervServices,
		secrets:      secretsStore,
		jobsEngine:   jobsEngine,
	}
}
