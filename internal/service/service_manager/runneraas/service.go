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
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers/github"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// RunneraasService implements service.RunnersService - see runners_api.proto's
// RunnersAPI doc comment. A runner is a normal Velez service, deployed
// through the ordinary VervServicesService.CreateNewDeploy path; only
// provider/scope/target/labels/secret_ref get their own linked row,
// velez.runners - status and environment are always read back through
// VervServicesService, never duplicated. Provider-specific behavior (minting
// a registration token, its env-var vocabulary) is delegated to a Provider,
// selected by RunnerProvider - everything else here is git-agnostic.
type RunneraasService struct {
	dataStorage  storage.Storage
	vervServices service.VervServicesService
	secrets      secrets.Store
	providers    map[velez_api.RunnerProvider]Provider
}

// New builds a RunneraasService. dataStorage, vervServices and secretsStore
// are interfaces, never concrete types - dataStorage is resolved per call
// (mirroring pgaas.PgaasService) so a runtime storage backend swap is picked
// up immediately.
func New(
	dataStorage storage.Storage, vervServices service.VervServicesService, secretsStore secrets.Store,
) *RunneraasService {
	providers := map[velez_api.RunnerProvider]Provider{
		velez_api.RunnerProvider_GITHUB: github.New(),
	}

	return &RunneraasService{
		dataStorage:  dataStorage,
		vervServices: vervServices,
		secrets:      secretsStore,
		providers:    providers,
	}
}

// boxes mirrors pgaas.PgaasService.boxes() - resolved per call so a runtime
// storage swap is picked up immediately.
func (s *RunneraasService) boxes() storage.ResourceBoxesStorage {
	return s.dataStorage.ResourceBoxes()
}

// provider looks up the Provider registered for p, wrapped so every caller
// gets the same rerrors-wrapped sentinel for an unrecognized/unspecified
// RunnerProvider.
func (s *RunneraasService) provider(p velez_api.RunnerProvider) (Provider, error) {
	found, ok := s.providers[p]
	if !ok {
		return nil, rerrors.Wrap(user_errors.ErrRunnerProviderUnsupported)
	}

	return found, nil
}
