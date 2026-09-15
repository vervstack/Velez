package registryaas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
)

// CreateRegistryInstance enqueues the multi-step create_registry_instance
// task (credentials, port resolution, htpasswd loader, the registry and its
// optional UI sidecar deploy, both storage rows -
// internal/jobs/create_registry_instance.go) and returns immediately -
// unlike ListPgInstances/CreatePgInstance, this can't be done inline:
// creating an instance means deploying containers through the ordinary
// (asynchronous) CreateNewDeploy/deploy watcher path, not something a single
// service-layer call can do by itself. Callers watch progress through
// TasksApi.WatchTask(req.Name, jobs.CreateRegistryInstanceAction) and refetch
// ListRegistryInstances once the task reaches DONE.
func (s *RegistryaasService) CreateRegistryInstance(ctx context.Context, req domain.CreateRegistryInstanceReq) error {
	initialContext := &velez_api.CreateRegistryInstanceTaskPayload{
		Request: registryInstanceRequestToPb(req),
	}

	_, err := s.jobsEngine.Enqueue(ctx, req.Name, jobs.CreateRegistryInstanceAction, initialContext)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing create registry instance task")
	}

	return nil
}

// registryInstanceRequestToPb converts the domain request into the wire
// request the create_registry_instance task persists as its own context -
// internal/jobs/create_registry_instance.go reads request fields straight off
// *velez_api.CreateRegistryInstance_Request, so the task payload carries the
// proto message rather than a second, parallel domain-shaped copy.
func registryInstanceRequestToPb(req domain.CreateRegistryInstanceReq) *velez_api.CreateRegistryInstance_Request {
	pbReq := &velez_api.CreateRegistryInstance_Request{
		Name:     req.Name,
		EnableUi: req.EnableUi,
	}

	if req.Environment != "" {
		pbReq.Environment = &req.Environment
	}

	if req.Box != "" {
		pbReq.Box = &req.Box
	}

	if req.ExposeToPort != 0 {
		pbReq.ExposeToPort = &req.ExposeToPort
	}

	if req.OwnerService != "" {
		pbReq.OwnerService = &req.OwnerService
	}

	return pbReq
}
